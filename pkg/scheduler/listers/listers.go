package listers

import (
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
	"k8s.io/kubernetes/pkg/scheduler/types"
)

// StackFilter is a function to filter a stack. If stack passed return true else return false.
type StackFilter func(stack *types.Stack) bool

// StackLister interface represents anything that can list stacks for a scheduler.
type StackLister interface {
	// Returns the list of pods.
	List() ([]*types.Stack, error)
	// This is similar to "List()", but the returned slice does not
	// contain pods that don't pass `podFilter`.
	FilteredList(stackFilter StackFilter) ([]*types.Stack, error)
}

// NodeInfoLister interface represents anything that can list/get NodeInfo objects from node name.
type NodeInfoLister interface {
	// Returns the list of NodeInfos.
	List() ([]*schedulernodeinfo.NodeInfo, error)
	// Returns the list of NodeInfos of nodes with pods with affinity terms.
	HavePodsWithAffinityList() ([]*schedulernodeinfo.NodeInfo, error)
	// Returns the NodeInfo of the given node name.
	Get(nodeName string) (*schedulernodeinfo.NodeInfo, error)
}

// SharedLister groups scheduler-specific listers.
type SharedLister interface {
	Stacks() StackLister
	NodeInfos() NodeInfoLister
}
