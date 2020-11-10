package flavor

import (
	"context"
	"fmt"
	"math"

	"k8s.io/kubernetes/pkg/scheduler/common/config"
	"k8s.io/kubernetes/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
	"k8s.io/kubernetes/pkg/scheduler/client/typed"
)

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "Flavor"

	// preFilterStateKey is the key in CycleState to Flavor pre-computed data.
	// Using the name of the plugin will likely help us avoid collisions with other plugins.
	preFilterStateKey = "PreFilter" + Name
)

// BaseWeight is weight
const BaseWeight = 10

// Flavor is a plugin that implements Priority based sorting.
type Flavor struct {
	handle interfaces.FrameworkHandle
}

var _ interfaces.ScorePlugin = &Flavor{}
var _ interfaces.PreFilterPlugin = &Flavor{}
var _ interfaces.FilterPlugin = &Flavor{}

type FlavorSlice []typed.Flavor

// flavorSlice to string
func (fs FlavorSlice) ToString() string {
	var ret = ""
	for _, fla := range fs {
		if ret != "" {
			ret += ","
		}
		ret += fla.ID
	}
	return ret
}

// Name returns name of the plugin.
func (pl *Flavor) Name() string {
	return Name
}

type FlavorAggregate struct {
	Flavors []interfaces.FlavorInfo
	Weight  int
}

type FlavorAggregates struct {
	Aggregates []FlavorAggregate
}

// Clone the prefilter state.
func (s *FlavorAggregates) Clone() interfaces.StateData {
	return s
}

func nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}

func calculateFlavorAggregate(stack *types.Stack) *FlavorAggregates {
	var ret = &FlavorAggregates{}

	lens := func(i int) int { return len(stack.Resources[i].Flavors) }

	for ix := make([]int, len(stack.Resources)); ix[0] < lens(0); nextIndex(ix, lens) {
		var temp = FlavorAggregate{Weight: 0}
		for j, k := range ix {
			flavorInfo := interfaces.FlavorInfo{Flavor: stack.Resources[j].Flavors[k], Count: stack.Resources[j].Count}
			temp.Flavors = append(temp.Flavors, flavorInfo)
			temp.Weight += BaseWeight / (k + 1)
		}
		ret.Aggregates = append(ret.Aggregates, temp)
	}
	return ret
}

func GetPreFilterState(cycleState *interfaces.CycleState) (*FlavorAggregates, error) {
	cycleState.RLock()
	defer cycleState.RUnlock()
	c, err := cycleState.Read(preFilterStateKey)
	if err != nil {
		logger.Errorf("error reading %q from cycleState: %s", preFilterStateKey, err)
		// NodeResFilterState doesn't exist, likely PreFilter wasn't invoked.
		return nil, fmt.Errorf("error reading %q from cycleState: %v", preFilterStateKey, err)
	}

	s, ok := c.(*FlavorAggregates)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to FlavorAggregates.preFilterState error", c)
	}
	return s, nil
}

// PreFilter invoked at the prefilter extension point.
func (f *Flavor) PreFilter(ctx context.Context, cycleState *interfaces.CycleState,
	stack *types.Stack) *interfaces.Status {
	cycleState.Lock()
	defer cycleState.Unlock()
	cycleState.Write(preFilterStateKey, calculateFlavorAggregate(stack))
	return nil
}

type spotHours struct {
	IsSpotBlock               bool
	InterruptionPolicy        string
	Infinite                  int
	PreemptibleResourceFinite map[string]int
}

func getFlavorToCount(flvAg FlavorAggregate) (map[string]int, map[string]*spotHours) {
	var flavorMap = map[string]int{}
	var spotFlavorMap = map[string]*spotHours{}
	for _, flv := range flvAg.Flavors {
		if flv.Spot != nil {
			if _, ok := spotFlavorMap[flv.FlavorID]; !ok {
				spotFlavorMap[flv.FlavorID] = &spotHours{PreemptibleResourceFinite: make(map[string]int)}
			}

			if flv.Spot.SpotDurationHours > 0 {
				(spotFlavorMap[flv.FlavorID]).IsSpotBlock = true
				(spotFlavorMap[flv.FlavorID]).InterruptionPolicy = flv.Spot.InterruptionPolicy
				hourCount := flv.Spot.SpotDurationHours * flv.Spot.SpotDurationCount
				var key = fmt.Sprintf("%dh", hourCount)

				prFinite := (spotFlavorMap[flv.FlavorID]).PreemptibleResourceFinite
				if _, ok := prFinite[key]; !ok {
					prFinite[key] = 0
				}

				prFinite[key] += flv.Count
			} else {
				(spotFlavorMap[flv.FlavorID]).IsSpotBlock = false
				(spotFlavorMap[flv.FlavorID]).InterruptionPolicy = flv.Spot.InterruptionPolicy
				(spotFlavorMap[flv.FlavorID]).Infinite += flv.Count
			}

		} else {
			if _, ok := flavorMap[flv.FlavorID]; ok {
				flavorMap[flv.FlavorID] = 0
			}
			flavorMap[flv.FlavorID] += flv.Count
		}
	}

	return flavorMap, spotFlavorMap
}

func isComFlavorMatch(ctx context.Context, flavorMap map[string]int, nodeInfo *nodeinfo.NodeInfo) (bool, int) {

	var maxCount = math.MaxFloat64

	for flavorID, count := range flavorMap {
		totalCount, ok := nodeInfo.AllocatableFlavor[flavorID]
		if !ok {
			logger.Debug(ctx, "Node (%s-%s) do not support required flavor (%s).",
				nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID)
			return false, 0
		}

		if totalCount < int64(count) {
			logger.Debug(ctx, "Node (%s-%s) flavor (%s) support Count(%d), required Count(%d).",
				nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, totalCount, count)
			return false, 0
		}

		maxCount = math.Min(maxCount, float64(totalCount/int64(count)))
	}

	return true, int(maxCount)
}

func isSpotFlavorMatch(ctx context.Context, spotFlavorMap map[string]*spotHours,
	nodeInfo *nodeinfo.NodeInfo) (bool, int) {
	if config.DefaultBool("spot_fake", false) {
		return true, 1000
	}

	var maxCount = math.MaxFloat64

	for flavorID, requestSpot := range spotFlavorMap {
		spotFlv, ok := nodeInfo.AllocatableSpotFlavor[flavorID]
		if !ok {
			logger.Debug(ctx, "Node (%s-%s) do not support required spot flavor (%s).",
				nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID)
			return false, 0
		}

		if requestSpot.IsSpotBlock {
			for key, value := range requestSpot.PreemptibleResourceFinite {
				var cpuResource types.PreemptibleResource
				var memResource types.PreemptibleResource
				if requestSpot.InterruptionPolicy == constants.InterruptionPolicyDelay {
					cpuResource = spotFlv.CpuBlockDelayResource
					memResource = spotFlv.MemoryBlockDelayResource
				} else {
					cpuResource = spotFlv.CpuBlockImmediateResource
					memResource = spotFlv.MemoryBlockImmediateResource
				}
				cpuHourValue, ok := cpuResource.Finite[key]
				if !ok {
					logger.Debug(ctx, "Node (%s-%s),flavor(%s), spot hours key(%s) not exist.",
						nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, key)
					return false, 0
				}

				if cpuHourValue < float32(value) {
					logger.Debug(ctx, "Node (%s-%s),flavor(%s), spot hours key(%s), "+
						"totalCount: %d, requestCount: %d.",
						nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, key, cpuHourValue, value)
					return false, 0
				}

				maxCount = math.Min(maxCount, float64(cpuHourValue/float32(value)))

				memHourValue, ok := memResource.Finite[key]
				if !ok {
					logger.Debug(ctx, "Node (%s-%s),flavor(%s), spot hours key(%s) not exist.",
						nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, key)
					return false, 0
				}

				if memHourValue < float32(value) {
					logger.Debug(ctx, "Node (%s-%s),flavor(%s), spot hours key(%s), "+
						"totalCount: %d, requestCount: %d.",
						nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, key, memHourValue, value)
					return false, 0
				}

				maxCount = math.Min(maxCount, float64(memHourValue/float32(value)))
			}
		} else {
			var cpuResource types.PreemptibleResource
			var memResource types.PreemptibleResource
			if requestSpot.InterruptionPolicy == constants.InterruptionPolicyDelay {
				cpuResource = spotFlv.CpuBlockDelayResource
				memResource = spotFlv.MemoryBlockDelayResource
			} else {
				cpuResource = spotFlv.CpuBlockImmediateResource
				memResource = spotFlv.MemoryBlockImmediateResource
			}
			if float32(requestSpot.Infinite) < cpuResource.Infinite {
				logger.Debug(ctx, "Node (%s-%s),flavor(%s), totalCount: %d, requestCount: %d.",
					nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, cpuResource.Infinite, requestSpot.Infinite)
				return false, 0
			}
			maxCount = math.Min(maxCount, float64(float32(requestSpot.Infinite)/cpuResource.Infinite))

			if float32(requestSpot.Infinite) < memResource.Infinite {
				logger.Debug(ctx, "Node (%s-%s),flavor(%s), totalCount: %d, requestCount: %d.",
					nodeInfo.Node().SiteID, nodeInfo.Node().Region, flavorID, memResource.Infinite, requestSpot.Infinite)
				return false, 0
			}

			maxCount = math.Min(maxCount, float64(float32(requestSpot.Infinite)/cpuResource.Infinite))
		}
	}

	return true, int(maxCount)
}

func (f *Flavor) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	nodeInfo *nodeinfo.NodeInfo) *interfaces.Status {
	s, err := GetPreFilterState(cycleState)
	if err != nil {
		return interfaces.NewStatus(interfaces.Error, err.Error())
	}

	for _, agg := range s.Aggregates {
		flavorMap, spotFlavorMap := getFlavorToCount(agg)

		var isCommonMatch, _ = isComFlavorMatch(ctx, flavorMap, nodeInfo)
		var isSpotMatch, _ = isSpotFlavorMatch(ctx, spotFlavorMap, nodeInfo)
		if isCommonMatch && isSpotMatch {
			return nil
		}
	}

	msg := fmt.Sprintf("Node (%s-%s) do not support required flavor (%+v).",
		nodeInfo.Node().SiteID, nodeInfo.Node().Region, s.Aggregates)
	logger.Info(ctx, msg)
	return interfaces.NewStatus(interfaces.Unschedulable, msg)
}

// Score invoked at the score extension point.
func (pl *Flavor) Score(ctx context.Context, state *interfaces.CycleState, stack *types.Stack,
	nodeID string) (int64, *interfaces.Status) {
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeID)
	if err != nil {
		return 0, interfaces.NewStatus(interfaces.Error, fmt.Sprintf("getting node %q from Snapshot: %v",
			nodeID, err))
	}

	s, err := GetPreFilterState(state)
	if err != nil {
		return 0, interfaces.NewStatus(interfaces.Error, err.Error())
	}

	for _, agg := range s.Aggregates {
		flavorMap, spotFlavorMap := getFlavorToCount(agg)

		var maxCount = 0
		var isCommonMatch, maxComCount = isComFlavorMatch(ctx, flavorMap, nodeInfo)
		var isSpotMatch, maxSpotCount = isSpotFlavorMatch(ctx, spotFlavorMap, nodeInfo)
		if isCommonMatch && isSpotMatch {

			if len(flavorMap) > 0 && len(spotFlavorMap) > 0 {
				maxCount = int(math.Min(float64(maxComCount), float64(maxSpotCount)))
			} else if len(flavorMap) > 0 {
				maxCount = maxComCount
			} else if len(spotFlavorMap) > 0 {
				maxCount = maxSpotCount
			}

			interfaces.UpdateNodeSelectorState(state, nodeID,
				map[string]interface{}{"Flavors": agg.Flavors, "StackMaxCount": maxCount})

			var score int64 = 0
			if len(agg.Flavors) > 0 {
				score = int64(agg.Weight / len(agg.Flavors))
			}
			return score, nil
		}
	}

	return 0, nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &Flavor{handle: handle}, nil
}
