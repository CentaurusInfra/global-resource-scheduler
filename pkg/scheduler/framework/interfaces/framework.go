package interfaces

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	internalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"
	schedulerlisters "k8s.io/kubernetes/pkg/scheduler/listers"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
	"k8s.io/kubernetes/pkg/scheduler/utils"
	"k8s.io/kubernetes/pkg/scheduler/utils/sets"
	"k8s.io/kubernetes/pkg/scheduler/utils/workqueue"
)

const (
	// Filter is the name of the filter extension point.
	Filter = "Filter"
)

// framework is the component responsible for initializing and running scheduler
// plugins.
type framework struct {
	registry              Registry
	snapshotSharedLister  schedulerlisters.SharedLister
	cache                 internalcache.Cache
	pluginNameToWeightMap map[string]int
	queueSortPlugins      []QueueSortPlugin
	preFilterPlugins      []PreFilterPlugin
	filterPlugins         []FilterPlugin
	preScorePlugins       []PreScorePlugin
	scorePlugins          []ScorePlugin
	reservePlugins        []ReservePlugin
	preBindPlugins        []PreBindPlugin
	bindPlugins           []BindPlugin
	postBindPlugins       []PostBindPlugin
	unreservePlugins      []UnreservePlugin
	permitPlugins         []PermitPlugin
	strategyPlugins       []StrategyPlugin

	// Indicates that RunFilterPlugins should accumulate all failed statuses and not return
	// after the first failure.
	runAllFilters bool
}

type frameworkOptions struct {
	snapshotSharedLister schedulerlisters.SharedLister
	cache                internalcache.Cache
	runAllFilters        bool
}

// Option for the framework.
type Option func(*frameworkOptions)

// WithSnapshotSharedLister sets the SharedLister of the snapshot.
func WithSnapshotSharedLister(snapshotSharedLister schedulerlisters.SharedLister) Option {
	return func(o *frameworkOptions) {
		o.snapshotSharedLister = snapshotSharedLister
	}
}

// WithCache sets the cache.
func WithCache(cache internalcache.Cache) Option {
	return func(o *frameworkOptions) {
		o.cache = cache
	}
}

// WithRunAllFilters sets the runAllFilters flag, which means RunFilterPlugins accumulates
// all failure Statuses.
func WithRunAllFilters(runAllFilters bool) Option {
	return func(o *frameworkOptions) {
		o.runAllFilters = runAllFilters
	}
}

func updatePluginList(pluginList interface{}, pluginSet *types.PluginSet, pluginsMap map[string]Plugin) error {
	if pluginSet == nil {
		return nil
	}

	plugins := reflect.ValueOf(pluginList).Elem()
	pluginType := plugins.Type().Elem()
	set := sets.NewString()
	for _, ep := range pluginSet.Enabled {
		pg, ok := pluginsMap[ep.Name]
		if !ok {
			return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
		}

		if !reflect.TypeOf(pg).Implements(pluginType) {
			return fmt.Errorf("plugin %q does not extend %s plugin", ep.Name, pluginType.Name())
		}

		if set.Has(ep.Name) {
			return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
		}

		set.Insert(ep.Name)

		newPlugins := reflect.Append(plugins, reflect.ValueOf(pg))
		plugins.Set(newPlugins)
	}
	return nil
}

var _ Framework = &framework{}

// NewFramework initializes plugins given the configuration and the registry.
func NewFramework(r Registry, plugins *types.Plugins, opts ...Option) (Framework, error) {
	options := frameworkOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	f := &framework{
		registry:              r,
		snapshotSharedLister:  options.snapshotSharedLister,
		cache:                 options.cache,
		pluginNameToWeightMap: make(map[string]int),
		runAllFilters:         options.runAllFilters,
	}
	if plugins == nil {
		return f, nil
	}

	// get needed plugins from config
	pg := f.pluginsNeeded(plugins)

	pluginsMap := make(map[string]Plugin)
	var totalPriority int64
	for name, factory := range r {
		// initialize only needed plugins.
		if _, ok := pg[name]; !ok {
			continue
		}

		p, err := factory(f)
		if err != nil {
			return nil, fmt.Errorf("error initializing plugin %q: %v", name, err)
		}
		pluginsMap[name] = p

		// a weight of zero is not permitted, plugins can be disabled explicitly
		// when configured.
		f.pluginNameToWeightMap[name] = int(pg[name].Weight)
		if f.pluginNameToWeightMap[name] == 0 {
			f.pluginNameToWeightMap[name] = 1
		}
		// Checks totalPriority against MaxTotalScore to avoid overflow
		if int64(f.pluginNameToWeightMap[name])*MaxNodeScore > MaxTotalScore-totalPriority {
			return nil, fmt.Errorf("total score of Score plugins could overflow")
		}
		totalPriority += int64(f.pluginNameToWeightMap[name]) * MaxNodeScore
	}

	for _, e := range f.getExtensionPoints(plugins) {
		if err := updatePluginList(e.slicePtr, e.plugins, pluginsMap); err != nil {
			return nil, err
		}
	}

	// Verifying the score weights again since Plugin.Name() could return a different
	// value from the one used in the configuration.
	for _, scorePlugin := range f.scorePlugins {
		if f.pluginNameToWeightMap[scorePlugin.Name()] == 0 {
			return nil, fmt.Errorf("score plugin %q is not configured with weight", scorePlugin.Name())
		}
	}

	if len(f.queueSortPlugins) == 0 {
		return nil, fmt.Errorf("no queue sort plugin is enabled")
	}
	if len(f.queueSortPlugins) > 1 {
		return nil, fmt.Errorf("only one queue sort plugin can be enabled")
	}
	if len(f.bindPlugins) == 0 {
		return nil, fmt.Errorf("at least one bind plugin is needed")
	}

	return f, nil
}

// QueueSortFunc returns the function to sort pods in scheduling queue
func (f *framework) QueueSortFunc() LessFunc {
	if f == nil {
		// If framework is nil, simply keep their order unchanged.
		// NOTE: this is primarily for tests.
		return func(_, _ *StackInfo) bool { return false }
	}

	if len(f.queueSortPlugins) == 0 {
		panic("No QueueSort plugin is registered in the framework.")
	}

	// Only one QueueSort plugin can be enabled.
	return f.queueSortPlugins[0].Less
}

// RunPreFilterPlugins runs the set of configured PreFilter plugins. It returns
// *Status and its code is set to non-success if any of the plugins returns
// anything but Success. If a non-success status is returned, then the scheduling
// cycle is aborted.
func (f *framework) RunPreFilterPlugins(ctx context.Context, state *CycleState, stack *types.Stack) (status *Status) {

	for _, pl := range f.preFilterPlugins {
		status = f.runPreFilterPlugin(ctx, pl, state, stack)
		if !status.IsSuccess() {
			if status.IsUnschedulable() {
				msg := fmt.Sprintf("rejected by %q at prefilter: %v", pl.Name(), status.Message())
				logger.Infof(msg)
				return NewStatus(status.Code(), msg)
			}
			msg := fmt.Sprintf("error while running %q prefilter plugin for pod %q: %v", pl.Name(),
				stack.Name, status.Message())
			logger.Errorf(msg)
			return NewStatus(Error, msg)
		}
	}

	return nil
}

func (f *framework) runPreFilterPlugin(ctx context.Context, pl PreFilterPlugin, state *CycleState,
	stack *types.Stack) *Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreFilter(ctx, state, stack)
	}
	status := pl.PreFilter(ctx, state, stack)
	return status
}

// RunFilterPlugins runs the set of configured Filter plugins for pod on
// the given node. If any of these plugins doesn't return "Success", the
// given node is not suitable for running pod.
// Meanwhile, the failure message and status are set for the given node.
func (f *framework) RunFilterPlugins(
	ctx context.Context,
	state *CycleState,
	stack *types.Stack,
	nodeInfo *schedulernodeinfo.NodeInfo,
) PluginToStatus {
	var firstFailedStatus *Status
	statuses := make(PluginToStatus)
	for _, pl := range f.filterPlugins {
		pluginStatus := f.runFilterPlugin(ctx, pl, state, stack, nodeInfo)
		if len(statuses) == 0 {
			firstFailedStatus = pluginStatus
		}
		if !pluginStatus.IsSuccess() {
			if !pluginStatus.IsUnschedulable() {
				// Filter plugins are not supposed to return any status other than
				// Success or Unschedulable.
				firstFailedStatus = NewStatus(Error, fmt.Sprintf("running %q filter plugin for pod %q: %v",
					pl.Name(), stack.Name, pluginStatus.Message()))
				return map[string]*Status{pl.Name(): firstFailedStatus}
			}
			statuses[pl.Name()] = pluginStatus
			if !f.runAllFilters {
				// Exit early if we don't need to run all filters.
				return statuses
			}
		}
	}

	return statuses
}

func (f *framework) runFilterPlugin(ctx context.Context, pl FilterPlugin, state *CycleState, stack *types.Stack,
	nodeInfo *schedulernodeinfo.NodeInfo) *Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Filter(ctx, state, stack, nodeInfo)
	}

	status := pl.Filter(ctx, state, stack, nodeInfo)
	return status
}

// RunPreScorePlugins runs the set of configured pre-score plugins. If any
// of these plugins returns any status other than "Success", the given pod is rejected.
func (f *framework) RunPreScorePlugins(
	ctx context.Context,
	state *CycleState,
	stack *types.Stack,
	nodes []*types.SiteNode,
) (status *Status) {

	for _, pl := range f.preScorePlugins {
		status = f.runPreScorePlugin(ctx, pl, state, stack, nodes)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running %q prescore plugin for pod %q: %v", pl.Name(),
				stack.Name, status.Message())
			logger.Errorf(msg)
			return NewStatus(Error, msg)
		}
	}

	return nil
}

func (f *framework) runPreScorePlugin(ctx context.Context, pl PreScorePlugin, state *CycleState, stack *types.Stack,
	nodes []*types.SiteNode) *Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreScore(ctx, state, stack, nodes)
	}
	status := pl.PreScore(ctx, state, stack, nodes)
	return status
}

// RunScorePlugins runs the set of configured scoring plugins. It returns a list that
// stores for each scoring plugin name the corresponding NodeScoreList(s).
// It also returns *Status, which is set to non-success if any of the plugins returns
// a non-success status.
func (f *framework) RunScorePlugins(ctx context.Context, state *CycleState, stack *types.Stack,
	nodes []*types.SiteNode) (ps PluginToNodeScores, status *Status) {

	pluginToNodeScores := make(PluginToNodeScores, len(f.scorePlugins))
	for _, pl := range f.scorePlugins {
		pluginToNodeScores[pl.Name()] = make(NodeScoreList, len(nodes))
	}
	ctx, cancel := context.WithCancel(ctx)
	errCh := utils.NewErrorChannel()

	// Run Score method for each node in parallel.
	workqueue.ParallelizeUntil(ctx, 16, len(nodes), func(index int) {
		for _, pl := range f.scorePlugins {
			nodeName := nodes[index].SiteID
			s, status := f.runScorePlugin(ctx, pl, state, stack, nodeName)
			if !status.IsSuccess() {
				errCh.SendErrorWithCancel(fmt.Errorf(status.Message()), cancel)
				return
			}
			pluginToNodeScores[pl.Name()][index] = NodeScore{
				Name:  nodeName,
				Score: int64(s),
			}
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while running score plugin for pod %q: %v", stack.Name, err)
		logger.Errorf("%s", msg)
		return nil, NewStatus(Error, msg)
	}

	// Apply score defaultWeights for each ScorePlugin in parallel.
	workqueue.ParallelizeUntil(ctx, 16, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		// Score plugins' weight has been checked when they are initialized.
		weight := f.pluginNameToWeightMap[pl.Name()]
		nodeScoreList := pluginToNodeScores[pl.Name()]

		for i, nodeScore := range nodeScoreList {
			// return error if score plugin returns invalid score.
			if nodeScore.Score > int64(MaxNodeScore) || nodeScore.Score < int64(MinNodeScore) {
				err := fmt.Errorf("score plugin %q returns an invalid score %v"+
					", it should in the range of [%v, %v] after normalizing", pl.Name(),
					nodeScore.Score, MinNodeScore, MaxNodeScore)
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			nodeScoreList[i].Score = nodeScore.Score * int64(weight)
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while applying score defaultWeights for pod %q: %v", stack.Name, err)
		logger.Errorf("%s", msg)
		return nil, NewStatus(Error, msg)
	}

	return pluginToNodeScores, nil
}

func (f *framework) runScorePlugin(ctx context.Context, pl ScorePlugin, state *CycleState, stack *types.Stack,
	nodeName string) (int64, *Status) {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Score(ctx, state, stack, nodeName)
	}

	s, status := pl.Score(ctx, state, stack, nodeName)
	return s, status
}

// RunPreBindPlugins runs the set of configured prebind plugins. It returns a
// failure (bool) if any of the plugins returns an error. It also returns an
// error containing the rejection message or the error occurred in the plugin.
func (f *framework) RunPreBindPlugins(ctx context.Context, state *CycleState, stack *types.Stack,
	nodeName string) (status *Status) {

	for _, pl := range f.preBindPlugins {
		status = f.runPreBindPlugin(ctx, pl, state, stack, nodeName)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running %q prebind plugin for pod %q: %v", pl.Name(),
				stack.Name, status.Message())
			logger.Errorf("%s", msg)
			return NewStatus(Error, msg)
		}
	}
	return nil
}

func (f *framework) runPreBindPlugin(ctx context.Context, pl PreBindPlugin, state *CycleState, stack *types.Stack,
	nodeName string) *Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.PreBind(ctx, state, stack, nodeName)
	}

	status := pl.PreBind(ctx, state, stack, nodeName)

	return status
}

// RunBindPlugins runs the set of configured bind plugins until one returns a non `Skip` status.
func (f *framework) RunBindPlugins(ctx context.Context, state *CycleState, stack *types.Stack,
	nodeName string) (status *Status) {
	if len(f.bindPlugins) == 0 {
		return NewStatus(Skip, "")
	}
	for _, bp := range f.bindPlugins {
		status = f.runBindPlugin(ctx, bp, state, stack, nodeName)
		if status != nil && status.Code() == Skip {
			continue
		}
		if !status.IsSuccess() {
			msg := fmt.Sprintf("plugin %q failed to bind pod \"%v\": %v", bp.Name(), stack.Name, status.Message())
			logger.Errorf("%s", msg)
			return NewStatus(Error, msg)
		}
		return status
	}
	return status
}

func (f *framework) runBindPlugin(ctx context.Context, bp BindPlugin, state *CycleState, stack *types.Stack,
	nodeName string) *Status {
	if !state.ShouldRecordPluginMetrics() {
		return bp.Bind(ctx, state, stack, nodeName)
	}

	status := bp.Bind(ctx, state, stack, nodeName)
	return status
}

// RunPostBindPlugins runs the set of configured postbind plugins.
func (f *framework) RunPostBindPlugins(ctx context.Context, state *CycleState, stack *types.Stack, nodeName string) {
	for _, pl := range f.postBindPlugins {
		f.runPostBindPlugin(ctx, pl, state, stack, nodeName)
	}
}

func (f *framework) runPostBindPlugin(ctx context.Context, pl PostBindPlugin, state *CycleState,
	stack *types.Stack, nodeName string) {
	if !state.ShouldRecordPluginMetrics() {
		pl.PostBind(ctx, state, stack, nodeName)
		return
	}
	pl.PostBind(ctx, state, stack, nodeName)

}

// RunReservePlugins runs the set of configured reserve plugins. If any of these
// plugins returns an error, it does not continue running the remaining ones and
// returns the error. In such case, pod will not be scheduled.
func (f *framework) RunReservePlugins(ctx context.Context, state *CycleState, stack *types.Stack,
	nodeName string) (status *Status) {
	for _, pl := range f.reservePlugins {
		status = f.runReservePlugin(ctx, pl, state, stack, nodeName)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running %q reserve plugin for pod %q: %v", pl.Name(),
				stack.Name, status.Message())
			logger.Errorf(msg)
			return NewStatus(Error, msg)
		}
	}
	return nil
}

func (f *framework) runReservePlugin(ctx context.Context, pl ReservePlugin, state *CycleState, stack *types.Stack,
	nodeName string) *Status {
	if !state.ShouldRecordPluginMetrics() {
		return pl.Reserve(ctx, state, stack, nodeName)
	}
	status := pl.Reserve(ctx, state, stack, nodeName)
	return status
}

// RunUnreservePlugins runs the set of configured unreserve plugins.
func (f *framework) RunUnreservePlugins(ctx context.Context, state *CycleState, stack *types.Stack, nodeName string) {
	for _, pl := range f.unreservePlugins {
		f.runUnreservePlugin(ctx, pl, state, stack, nodeName)
	}
}

func (f *framework) runUnreservePlugin(ctx context.Context, pl UnreservePlugin, state *CycleState,
	stack *types.Stack, nodeName string) {
	if !state.ShouldRecordPluginMetrics() {
		pl.Unreserve(ctx, state, stack, nodeName)
		return
	}

	pl.Unreserve(ctx, state, stack, nodeName)
}

// RunPermitPlugins runs the set of configured permit plugins. If any of these
// plugins returns a status other than "Success" or "Wait", it does not continue
// running the remaining plugins and returns an error. Otherwise, if any of the
// plugins returns "Wait", then this function will create and add waiting pod
// to a map of currently waiting pods and return status with "Wait" code.
// Pod will remain waiting pod for the minimum duration returned by the permit plugins.
func (f *framework) RunPermitPlugins(ctx context.Context, state *CycleState, stack *types.Stack,
	nodeName string) (status *Status) {
	return nil
}

func (f *framework) runPermitPlugin(ctx context.Context, pl PermitPlugin, state *CycleState, stack *types.Stack,
	nodeName string) (*Status, time.Duration) {
	return nil, time.Duration(1 * time.Second)
}

//RunStrategyPlugins runs the set of configured strategy plugins
func (f *framework) RunStrategyPlugins(ctx context.Context, state *CycleState,
	allocations *types.Allocation, nodeList NodeScoreList) (NodeScoreList, *Status) {

	countList := nodeList

	for _, pl := range f.strategyPlugins {
		var status *Status
		countList, status = pl.Strategy(ctx, state, allocations, countList)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("plugin %q failed to Strategy \"%v\": %v", pl.Name(),
				allocations.Stack.Name, status.Message())
			logger.Errorf("%s", msg)
			return nil, NewStatus(Error, msg)
		}
	}

	return countList, nil
}

// SnapshotSharedLister returns the scheduler's SharedLister of the latest NodeInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a pod finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *framework) SnapshotSharedLister() schedulerlisters.SharedLister {
	return f.snapshotSharedLister
}

// SnapshotSharedLister returns the scheduler's SharedLister of the latest NodeInfo
// snapshot. The snapshot is taken at the beginning of a scheduling cycle and remains
// unchanged until a pod finishes "Reserve". There is no guarantee that the information
// remains unchanged after "Reserve".
func (f *framework) Cache() internalcache.Cache {
	return f.cache
}

// HasFilterPlugins returns true if at least one filter plugin is defined.
func (f *framework) HasFilterPlugins() bool {
	return len(f.filterPlugins) > 0
}

// HasScorePlugins returns true if at least one score plugin is defined.
func (f *framework) HasScorePlugins() bool {
	return len(f.scorePlugins) > 0
}

// extensionPoint encapsulates desired and applied set of plugins at a specific extension
// point. This is used to simplify iterating over all extension points supported by the
// framework.
type extensionPoint struct {
	// the set of plugins to be configured at this extension point.
	plugins *types.PluginSet
	// a pointer to the slice storing plugins implementations that will run at this
	// extension point.
	slicePtr interface{}
}

func (f *framework) getExtensionPoints(plugins *types.Plugins) []extensionPoint {
	return []extensionPoint{
		{plugins.PreFilter, &f.preFilterPlugins},
		{plugins.Filter, &f.filterPlugins},
		{plugins.Reserve, &f.reservePlugins},
		{plugins.PreScore, &f.preScorePlugins},
		{plugins.Score, &f.scorePlugins},
		{plugins.PreBind, &f.preBindPlugins},
		{plugins.Bind, &f.bindPlugins},
		{plugins.PostBind, &f.postBindPlugins},
		{plugins.Unreserve, &f.unreservePlugins},
		{plugins.Permit, &f.permitPlugins},
		{plugins.QueueSort, &f.queueSortPlugins},
		{plugins.Strategy, &f.strategyPlugins},
	}
}

func (f *framework) pluginsNeeded(plugins *types.Plugins) map[string]types.Plugin {
	pgMap := make(map[string]types.Plugin)

	if plugins == nil {
		return pgMap
	}

	find := func(pgs *types.PluginSet) {
		if pgs == nil {
			return
		}
		for _, pg := range pgs.Enabled {
			pgMap[pg.Name] = pg
		}
	}
	for _, e := range f.getExtensionPoints(plugins) {
		find(e.plugins)
	}
	return pgMap
}
