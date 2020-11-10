package exclusivenode

import (
	"context"
	"strings"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/sets"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "ExclusiveNode"

// PrioritySort is a plugin that implements Priority based sorting.
type ExclusiveNode struct{}

var _ interfaces.FilterPlugin = &ExclusiveNode{}

// Name returns name of the plugin.
func (pl *ExclusiveNode) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *ExclusiveNode) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	nodeInfo *nodeinfo.NodeInfo) *interfaces.Status {
	domainID := utils.GetStrFromCtx(ctx, constants.ContextDomainID)

	siteAttrs := nodeInfo.Node().SiteAttribute

	for _, siteAttr := range siteAttrs {
		if siteAttr.Key == constants.SiteExclusiveDomains {
			whiteList := siteAttr.Value
			allowDomains := strings.Split(whiteList, ",")
			allowDomainsSet := sets.NewString(allowDomains...)
			if allowDomainsSet.Has(domainID) {
				logger.Debug(ctx, "Node(%s-%s) belong to domianID(%s)", nodeInfo.Node().SiteID,
					nodeInfo.Node().Region, domainID)
				return interfaces.NewStatus(interfaces.Unschedulable, "node is exclusive node.")
			}

			break
		}
	}

	return nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &ExclusiveNode{}, nil
}
