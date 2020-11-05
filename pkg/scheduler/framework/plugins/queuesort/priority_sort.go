package queuesort

import (
	"k8s.io/kubernetes/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/pkg/scheduler/types"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "PrioritySort"

func GetStackPriority(stack *types.Stack) int {
	return 1
}

// PrioritySort is a plugin that implements Priority based sorting.
type PrioritySort struct{}

var _ interfaces.QueueSortPlugin = &PrioritySort{}

// Name returns name of the plugin.
func (pl *PrioritySort) Name() string {
	return Name
}

// Less is the function used by the activeQ heap algorithm to sort pods.
// It sorts pods based on their priority. When priorities are equal, it uses
// PodInfo.timestamp.
func (pl *PrioritySort) Less(sInfo1, sInfo2 *interfaces.StackInfo) bool {
	p1 := GetStackPriority(sInfo1.Stack)
	p2 := GetStackPriority(sInfo2.Stack)
	return (p1 > p2) || (p1 == p2 && sInfo1.Timestamp.Before(sInfo2.Timestamp))
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &PrioritySort{}, nil
}
