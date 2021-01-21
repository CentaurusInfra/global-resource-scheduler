/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package interfaces

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	internalcache "k8s.io/kubernetes/globalscheduler/pkg/scheduler/internal/cache"
	schedulerlisters "k8s.io/kubernetes/globalscheduler/pkg/scheduler/listers"
	schedulersitecacheinfo "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
)

// SiteScoreList declares a list of sites and their scores.
type SiteScoreList []SiteScore

// SiteScore is a struct with sites name and score.
type SiteScore struct {
	SiteID     string
	Region     string
	AZ         string
	Score      int64
	StackCount int
}

func (f SiteScoreList) Len() int { return len(f) }

func (f SiteScoreList) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func (f SiteScoreList) Less(i, j int) bool {
	return f[i].Score < f[j].Score
}

// PluginToSiteScores declares a map from plugin name to its SiteScoreList.
type PluginToSiteScores map[string]SiteScoreList

// SiteToStatusMap declares map from siteID to its status.
type SiteToStatusMap map[string]*Status

// ToString returns string
func (s SiteToStatusMap) ToString() string {
	var retStr = ""
	for key, value := range s {
		retStr += fmt.Sprintf("%s:[%s]", key, value.Message())
	}
	return retStr
}

// Code is the Status code/type which is returned from plugins.
type Code int

// These are predefined codes used in a Status.
const (
	// Success means that plugin ran correctly and found pod schedulable.
	// NOTE: A nil status is also considered as "Success".
	Success Code = iota
	// Error is used for internal plugin errors, unexpected input, etc.
	Error
	// Unschedulable is used when a plugin finds a pod unschedulable. The scheduler might attempt to
	// preempt other pods to get this pod scheduled. Use UnschedulableAndUnresolvable to make the
	// scheduler skip preemption.
	// The accompanying status message should explain why the pod is unschedulable.
	Unschedulable
	// UnschedulableAndUnresolvable is used when a (pre-)filter plugin finds a pod unschedulable and
	// preemption would not change anything. Plugins should return Unschedulable if it is possible
	// that the pod can get scheduled with preemption.
	// The accompanying status message should explain why the pod is unschedulable.
	UnschedulableAndUnresolvable
	// Wait is used when a permit plugin finds a pod scheduling should wait.
	Wait
	// Skip is used when a bind plugin chooses to skip binding.
	Skip
)

// This list should be exactly the same as the codes iota defined above in the same order.
var codes = []string{"Success", "Error", "Unschedulable", "UnschedulableAndUnresolvable", "Wait", "Skip"}

func (c Code) String() string {
	return codes[c]
}

const (
	// MaxSiteScore is the maximum score a Score plugin is expected to return.
	MaxSiteScore int64 = 100

	// MinSiteScore is the minimum score a Score plugin is expected to return.
	MinSiteScore int64 = 0

	// MaxTotalScore is the maximum total score.
	MaxTotalScore int64 = math.MaxInt64
)

// Status indicates the result of running a plugin. It consists of a code and a
// message. When the status code is not `Success`, the reasons should explain why.
// NOTE: A nil Status is also considered as Success.
type Status struct {
	code    Code
	reasons []string
}

// Code returns code of the Status.
func (s *Status) Code() Code {
	if s == nil {
		return Success
	}
	return s.code
}

// Message returns a concatenated message on reasons of the Status.
func (s *Status) Message() string {
	if s == nil {
		return ""
	}
	return strings.Join(s.reasons, ", ")
}

// Reasons returns reasons of the Status.
func (s *Status) Reasons() []string {
	return s.reasons
}

// AppendReason appends given reason to the Status.
func (s *Status) AppendReason(reason string) {
	s.reasons = append(s.reasons, reason)
}

// IsSuccess returns true if and only if "Status" is nil or Code is "Success".
func (s *Status) IsSuccess() bool {
	return s.Code() == Success
}

// IsUnschedulable returns true if "Status" is Unschedulable (Unschedulable or UnschedulableAndUnresolvable).
func (s *Status) IsUnschedulable() bool {
	code := s.Code()
	return code == Unschedulable || code == UnschedulableAndUnresolvable
}

// AsError returns nil if the status is a success; otherwise returns an "error" object
// with a concatenated message on reasons of the Status.
func (s *Status) AsError() error {
	if s.IsSuccess() {
		return nil
	}
	return errors.New(s.Message())
}

// NewStatus makes a Status out of the given arguments and returns its pointer.
func NewStatus(code Code, reasons ...string) *Status {
	return &Status{
		code:    code,
		reasons: reasons,
	}
}

// PluginToStatus maps plugin name to status. Currently used to identify which Filter plugin
// returned which status.
type PluginToStatus map[string]*Status

// Merge merges the statuses in the map into one. The resulting status code have the following
// precedence: Error, UnschedulableAndUnresolvable, Unschedulable.
func (p PluginToStatus) Merge() *Status {
	if len(p) == 0 {
		return nil
	}

	finalStatus := NewStatus(Success)
	var hasError, hasUnschedulableAndUnresolvable, hasUnschedulable bool
	for _, s := range p {
		if s.Code() == Error {
			hasError = true
		} else if s.Code() == UnschedulableAndUnresolvable {
			hasUnschedulableAndUnresolvable = true
		} else if s.Code() == Unschedulable {
			hasUnschedulable = true
		}
		finalStatus.code = s.Code()
		for _, r := range s.reasons {
			finalStatus.AppendReason(r)
		}
	}

	if hasError {
		finalStatus.code = Error
	} else if hasUnschedulableAndUnresolvable {
		finalStatus.code = UnschedulableAndUnresolvable
	} else if hasUnschedulable {
		finalStatus.code = Unschedulable
	}
	return finalStatus
}

// Plugin is the parent type for all the scheduling framework plugins.
type Plugin interface {
	Name() string
}

// StackInfo is a wrapper to a Stack with additional information for purposes such as tracking
// the timestamp when it's added to the queue or recording per-pod metrics.
type StackInfo struct {
	Stack *types.Stack
	// The time pod added to the scheduling queue.
	Timestamp time.Time
	// Number of schedule attempts before successfully scheduled.
	// It's used to record the # attempts metric.
	Attempts int
	// The time when the pod is added to the queue for the first time. The pod may be added
	// back to the queue multiple times before it's successfully scheduled.
	// It shouldn't be updated once initialized. It's used to record the e2e scheduling
	// latency for a pod.
	InitialAttemptTimestamp time.Time
}

// DeepCopy returns a deep copy of the PodInfo object.
func (stackInfo *StackInfo) DeepCopy() *StackInfo {
	return &StackInfo{
		Stack:                   stackInfo.Stack.DeepCopy(),
		Timestamp:               stackInfo.Timestamp,
		Attempts:                stackInfo.Attempts,
		InitialAttemptTimestamp: stackInfo.InitialAttemptTimestamp,
	}
}

// LessFunc is the function to sort pod info
type LessFunc func(stackInfo1, stackInfo2 *StackInfo) bool

// QueueSortPlugin is an interface that must be implemented by "QueueSort" plugins.
// These plugins are used to sort pods in the scheduling queue. Only one queue sort
// plugin may be enabled at a time.
type QueueSortPlugin interface {
	Plugin
	// Less are used to sort pods in the scheduling queue.
	Less(*StackInfo, *StackInfo) bool
}

// PreFilterPlugin is an interface that must be implemented by "prefilter" plugins.
// These plugins are called at the beginning of the scheduling cycle.
type PreFilterPlugin interface {
	Plugin
	// PreFilter is called at the beginning of the scheduling cycle. All PreFilter
	// plugins must return success or the pod will be rejected.
	PreFilter(ctx context.Context, state *CycleState, p *types.Stack) *Status
}

// FilterPlugin is an interface for Filter plugins. These plugins are called at the
// filter extension point for filtering out hosts that cannot run a pod.
// This concept used to be called 'predicate' in the original scheduler.
// These plugins should return "Success", "Unschedulable" or "Error" in Status.code.
// However, the scheduler accepts other valid codes as well.
// Anything other than "Success" will lead to exclusion of the given host from
// running the pod.
type FilterPlugin interface {
	Plugin
	// Filter is called by the scheduling framework.
	// All FilterPlugins should return "Success" to declare that
	// the given site fits the pod. If Filter doesn't return "Success",
	// please refer scheduler/algorithm/predicates/error.go
	// to set error message.
	// For the site being evaluated, Filter plugins should look at the passed
	// siteCacheInfo reference for this particular site's information (e.g., pods
	// considered to be running on the site) instead of looking it up in the
	// siteCacheInfoSnapshot because we don't guarantee that they will be the same.
	// For example, during preemption, we may pass a copy of the original
	// siteCacheInfo object that has some pods removed from it to evaluate the
	// possibility of preempting them to schedule the target pod.
	Filter(ctx context.Context, state *CycleState, stack *types.Stack, siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) *Status
}

// PreScorePlugin is an interface for Pre-score plugin. Pre-score is an
// informational extension point. Plugins will be called with a list of sites
// that passed the filtering phase. A plugin may use this data to update internal
// state or to generate logs/metrics.
type PreScorePlugin interface {
	Plugin
	// PreScore is called by the scheduling framework after a list of sites
	// passed the filtering phase. All prescore plugins must return success or
	// the pod will be rejected
	PreScore(ctx context.Context, state *CycleState, pod *types.Stack, sites []*types.Site) *Status
}

// ScorePlugin is an interface that must be implemented by "score" plugins to rank
// sites that passed the filtering phase.
type ScorePlugin interface {
	Plugin
	// Score is called on each filtered site. It must return success and an integer
	// indicating the rank of the site. All scoring plugins must return success or
	// the pod will be rejected.
	Score(ctx context.Context, state *CycleState, p *types.Stack, siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) (int64, *Status)
}

// PostScorePlugin is an interface that must be implemented by "PostScore" plugins to rank
// sites that passed the filtering phase.
type PostScorePlugin interface {
	Plugin
	// Score is called on each filtered site. It must return success and an integer
	// indicating the rank of the site. All scoring plugins must return success or
	// the pod will be rejected.
	PostScore(ctx context.Context, state *CycleState, p *types.Stack,
		siteScoreList SiteScoreList) (SiteScoreList, *Status)
}

// ReservePlugin is an interface for Reserve plugins. These plugins are called
// at the reservation point. These are meant to update the state of the plugin.
// This concept used to be called 'assume' in the original scheduler.
// These plugins should return only Success or Error in Status.code. However,
// the scheduler accepts other valid codes as well. Anything other than Success
// will lead to rejection of the pod.
type ReservePlugin interface {
	Plugin
	// Reserve is called by the scheduling framework when the scheduler cache is
	// updated.
	Reserve(ctx context.Context, state *CycleState, p *types.Stack, siteID string) *Status
}

// PreBindPlugin is an interface that must be implemented by "prebind" plugins.
// These plugins are called before a pod being scheduled.
type PreBindPlugin interface {
	Plugin
	// PreBind is called before binding a pod. All prebind plugins must return
	// success or the pod will be rejected and won't be sent for binding.
	PreBind(ctx context.Context, state *CycleState, p *types.Stack, siteID string) *Status
}

// PostBindPlugin is an interface that must be implemented by "postbind" plugins.
// These plugins are called after a pod is successfully bound to a site.
type PostBindPlugin interface {
	Plugin
	// PostBind is called after a pod is successfully bound. These plugins are
	// informational. A common application of this extension point is for cleaning
	// up. If a plugin needs to clean-up its state after a pod is scheduled and
	// bound, PostBind is the extension point that it should register.
	PostBind(ctx context.Context, state *CycleState, p *types.Stack, siteID string)
}

// UnreservePlugin is an interface for Unreserve plugins. This is an informational
// extension point. If a pod was reserved and then rejected in a later phase, then
// un-reserve plugins will be notified. Un-reserve plugins should clean up state
// associated with the reserved Pod.
type UnreservePlugin interface {
	Plugin
	// Unreserve is called by the scheduling framework when a reserved pod was
	// rejected in a later phase.
	Unreserve(ctx context.Context, state *CycleState, p *types.Stack, siteID string)
}

// PermitPlugin is an interface that must be implemented by "permit" plugins.
// These plugins are called before a pod is bound to a site.
type PermitPlugin interface {
	Plugin
	// Permit is called before binding a pod (and before prebind plugins). Permit
	// plugins are used to prevent or delay the binding of a Pod. A permit plugin
	// must return success or wait with timeout duration, or the pod will be rejected.
	// The pod will also be rejected if the wait timeout or the pod is rejected while
	// waiting. Note that if the plugin returns "wait", the framework will wait only
	// after running the remaining plugins given that no other plugin rejects the pod.
	Permit(ctx context.Context, state *CycleState, p *types.Stack, siteID string) (*Status, time.Duration)
}

// BindPlugin is an interface that must be implemented by "bind" plugins. Bind
// plugins are used to bind a pod to a Site.
type BindPlugin interface {
	Plugin
	// Bind plugins will not be called until all pre-bind plugins have completed. Each
	// bind plugin is called in the configured order. A bind plugin may choose whether
	// or not to handle the given Pod. If a bind plugin chooses to handle a Pod, the
	// remaining bind plugins are skipped. When a bind plugin does not handle a pod,
	// it must return Skip in its Status code. If a bind plugin returns an Error, the
	// pod is rejected and will not be bound.
	Bind(ctx context.Context, state *CycleState, p *types.Stack, siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) *Status
}

// StrategyPlugin is an interface that must be implemented by "strategy" plugins. strategy
// plugins are used to strategy.
type StrategyPlugin interface {
	Plugin

	//Strategy plugins will will not be called until all score plugins have completed
	Strategy(ctx context.Context, state *CycleState,
		allocations *types.Allocation, siteScoreList SiteScoreList) (SiteScoreList, *Status)
}

// Framework manages the set of plugins in use by the scheduling framework.
// Configured plugins are called at specified points in a scheduling context.
type Framework interface {
	FrameworkHandle
	// QueueSortFunc returns the function to sort pods in scheduling queue
	QueueSortFunc() LessFunc

	// RunPreFilterPlugins runs the set of configured prefilter plugins. It returns
	// *Status and its code is set to non-success if any of the plugins returns
	// anything but Success. If a non-success status is returned, then the scheduling
	// cycle is aborted.
	RunPreFilterPlugins(ctx context.Context, state *CycleState, stack *types.Stack) *Status

	// RunFilterPlugins runs the set of configured filter plugins for pod on
	// the given site. Note that for the site being evaluated, the passed site
	// reference could be different from the one in Snapshot map (e.g., pods
	// considered to be running on the site could be different). For example, during
	// preemption, we may pass a copy of the original site object that has some pods
	// removed from it to evaluate the possibility of preempting them to
	// schedule the target pod.
	RunFilterPlugins(ctx context.Context, state *CycleState, stack *types.Stack,
		siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) PluginToStatus

	// RunPreScorePlugins runs the set of configured pre-score plugins. If any
	// of these plugins returns any status other than "Success", the given pod is rejected.
	RunPreScorePlugins(ctx context.Context, state *CycleState, stack *types.Stack, sites []*types.Site) *Status

	// RunScorePlugins runs the set of configured scoring plugins. It returns a map that
	// stores for each scoring plugin name the corresponding SiteScoreList(s).
	// It also returns *Status, which is set to non-success if any of the plugins returns
	// a non-success status.
	RunScorePlugins(ctx context.Context, state *CycleState, stack *types.Stack,
		sites []*types.Site, siteCacheInfoMap map[string]*schedulersitecacheinfo.SiteCacheInfo) (PluginToSiteScores, *Status)

	// RunPreBindPlugins runs the set of configured prebind plugins. It returns
	// *Status and its code is set to non-success if any of the plugins returns
	// anything but Success. If the Status code is "Unschedulable", it is
	// considered as a scheduling check failure, otherwise, it is considered as an
	// internal error. In either case the pod is not going to be bound.
	RunPreBindPlugins(ctx context.Context, state *CycleState, stack *types.Stack, siteID string) *Status

	// RunReservePlugins runs the set of configured reserve plugins. If any of these
	// plugins returns an error, it does not continue running the remaining ones and
	// returns the error. In such case, pod will not be scheduled.
	RunReservePlugins(ctx context.Context, state *CycleState, stack *types.Stack, siteID string) *Status

	// RunUnreservePlugins runs the set of configured unreserve plugins.
	RunUnreservePlugins(ctx context.Context, state *CycleState, stack *types.Stack, siteID string)

	// RunPermitPlugins runs the set of configured permit plugins. If any of these
	// plugins returns a status other than "Success" or "Wait", it does not continue
	// running the remaining plugins and returns an error. Otherwise, if any of the
	// plugins returns "Wait", then this function will create and add waiting pod
	// to a map of currently waiting pods and return status with "Wait" code.
	// Pod will remain waiting pod for the minimum duration returned by the permit plugins.
	RunPermitPlugins(ctx context.Context, state *CycleState, stack *types.Stack, siteID string) *Status

	// RunBindPlugins runs the set of configured bind plugins. A bind plugin may choose
	// whether or not to handle the given Pod. If a bind plugin chooses to skip the
	// binding, it should return code=5("skip") status. Otherwise, it should return "Error"
	// or "Success". If none of the plugins handled binding, RunBindPlugins returns code=5("skip") status.
	RunBindPlugins(ctx context.Context, state *CycleState, stack *types.Stack,
		siteCacheInfo *schedulersitecacheinfo.SiteCacheInfo) *Status

	//RunStrategyPlugins runs the set of configured strategy plugins.
	RunStrategyPlugins(ctx context.Context, state *CycleState,
		allocations *types.Allocation, siteScoreList SiteScoreList) (SiteScoreList, *Status)

	// HasFilterPlugins returns true if at least one filter plugin is defined.
	HasFilterPlugins() bool

	// HasScorePlugins returns true if at least one score plugin is defined.
	HasScorePlugins() bool
}

// FrameworkHandle provides data and some tools that plugins can use. It is
// passed to the plugin factories at the time of plugin initialization. Plugins
// must store and use this handle to call framework functions.
type FrameworkHandle interface {
	// SnapshotSharedLister returns listers from the latest SiteCacheInfo Snapshot. The snapshot
	// is taken at the beginning of a scheduling cycle and remains unchanged until
	// a pod finishes "Permit" point. There is no guarantee that the information
	// remains unchanged in the binding phase of scheduling, so plugins in the binding
	// cycle (pre-bind/bind/post-bind/un-reserve plugin) should not use it,
	// otherwise a concurrent read/write error might occur, they should use scheduler
	// cache instead.
	SnapshotSharedLister() schedulerlisters.SharedLister

	Cache() internalcache.Cache
}
