package interfaces

import (
	"errors"
	"fmt"
	"sync"

	"k8s.io/kubernetes/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/pkg/scheduler/types"
)

const (
	// NotFound is the not found error message.
	NotFound = "not found"
)

//FlavorInfo flavor info
type FlavorInfo struct {
	types.Flavor
	Count int
}

// StateData is a generic type for arbitrary data stored in CycleState.
type StateData interface {
	// Clone is an interface to make a copy of StateData. For performance reasons,
	// clone should make shallow copies for members (e.g., slices or maps) that are not
	// impacted by PreFilter's optional AddPod/RemovePod methods.
	Clone() StateData
}

// StateKey is the type of keys stored in CycleState.
type StateKey string

// CycleState provides a mechanism for plugins to store and retrieve arbitrary data.
// StateData stored by one plugin can be read, altered, or deleted by another plugin.
// CycleState does not provide any data protection, as all plugins are assumed to be
// trusted.
type CycleState struct {
	mx      sync.RWMutex
	storage map[StateKey]StateData
	// if recordPluginMetrics is true, PluginExecutionDuration will be recorded for this cycle.
	recordPluginMetrics bool
}

// NewCycleState initializes a new CycleState and returns its pointer.
func NewCycleState() *CycleState {
	return &CycleState{
		storage: make(map[StateKey]StateData),
	}
}

// ShouldRecordPluginMetrics returns whether PluginExecutionDuration metrics should be recorded.
func (c *CycleState) ShouldRecordPluginMetrics() bool {
	if c == nil {
		return false
	}
	return c.recordPluginMetrics
}

// SetRecordPluginMetrics sets recordPluginMetrics to the given value.
func (c *CycleState) SetRecordPluginMetrics(flag bool) {
	if c == nil {
		return
	}
	c.recordPluginMetrics = flag
}

// Clone creates a copy of CycleState and returns its pointer. Clone returns
// nil if the context being cloned is nil.
func (c *CycleState) Clone() *CycleState {
	if c == nil {
		return nil
	}
	copy := NewCycleState()
	for k, v := range c.storage {
		copy.Write(k, v.Clone())
	}
	return copy
}

// Read retrieves data with the given "key" from CycleState. If the key is not
// present an error is returned.
// This function is not thread safe. In multi-threaded code, lock should be
// acquired first.
func (c *CycleState) Read(key StateKey) (StateData, error) {
	if v, ok := c.storage[key]; ok {
		if v != nil {
			return v, nil
		}

		return nil, nil
	}
	return nil, errors.New(NotFound)
}

// Write stores the given "val" in CycleState with the given "key".
// This function is not thread safe. In multi-threaded code, lock should be
// acquired first.
func (c *CycleState) Write(key StateKey, val StateData) {
	c.storage[key] = val
}

// Delete deletes data with the given key from CycleState.
// This function is not thread safe. In multi-threaded code, lock should be
// acquired first.
func (c *CycleState) Delete(key StateKey) {
	delete(c.storage, key)
}

// Lock acquires CycleState lock.
func (c *CycleState) Lock() {
	c.mx.Lock()
}

// Unlock releases CycleState lock.
func (c *CycleState) Unlock() {
	c.mx.Unlock()
}

// RLock acquires CycleState read lock.
func (c *CycleState) RLock() {
	c.mx.RLock()
}

// RUnlock releases CycleState read lock.
func (c *CycleState) RUnlock() {
	c.mx.RUnlock()
}

//NodeSelectorInfo node temp info
type NodeSelectorInfo struct {
	Flavors       []FlavorInfo
	VolumeType    []string
	StackMaxCount int
}

//NodeSelectorInfo node temp info
type NodeSelectorInfos map[string]NodeSelectorInfo

// Clone the prefilter state.
func (s *NodeSelectorInfos) Clone() StateData {
	return s
}

func UpdateNodeSelectorState(cycleState *CycleState, nodeID string, updateInfo map[string]interface{}) error {
	if updateInfo == nil {
		return nil
	}

	cycleState.Lock()
	defer cycleState.Unlock()
	c, err := cycleState.Read(types.NodeTempSelectorKey)
	if err != nil {
		selectInfos := NodeSelectorInfos{}
		selectInfo := NodeSelectorInfo{}

		selectInfos[nodeID] = selectInfo
		cycleState.Write(types.NodeTempSelectorKey, &selectInfos)
	}

	s, ok := c.(*NodeSelectorInfos)
	if !ok {
		return fmt.Errorf("%+v  convert to NodeSelectorInfos error", c)
	}

	selectInfo, ok := (*s)[nodeID]
	if !ok {
		selectInfo = NodeSelectorInfo{}
	}

	for key, value := range updateInfo {
		switch key {
		case "Flavors":
			flvs, ok := value.([]FlavorInfo)
			if !ok {
				return fmt.Errorf("%+v  convert to []string error", c)
			}
			selectInfo.Flavors = flvs
		case "StackMaxCount":
			count, ok := value.(int)
			if !ok {
				return fmt.Errorf("%+v  convert to int error", c)
			}

			if selectInfo.StackMaxCount == 0 || selectInfo.StackMaxCount > count {
				selectInfo.StackMaxCount = count
			}
		default:
			logger.Warnf("key = %s not support!", key)
		}
	}

	(*s)[nodeID] = selectInfo

	cycleState.Write(types.NodeTempSelectorKey, s)

	return nil
}

func GetNodeSelectorState(cycleState *CycleState, nodeID string) (*NodeSelectorInfo, error) {
	c, err := cycleState.Read(types.NodeTempSelectorKey)
	if err != nil {
		return nil, fmt.Errorf("error reading %q from cycleState: %v", types.NodeTempSelectorKey, err)
	}

	s, ok := c.(*NodeSelectorInfos)
	if !ok {
		return nil, fmt.Errorf("%+v  convert to NodeSelectorInfos error", c)
	}

	value, ok := (*s)[nodeID]
	if !ok {
		return nil, fmt.Errorf("nodeID(%s) not found", nodeID)
	}

	return &value, nil
}
