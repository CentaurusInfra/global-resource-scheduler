package cache

import (
	"fmt"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

// nodeTree is a tree-like data structure that holds node names in each zone. Zone names are
// keys to "NodeTree.tree" and values of "NodeTree.tree" are arrays of node names.
// NodeTree is NOT thread-safe, any concurrent updates/reads from it must be synchronized by the caller.
// It is used only by schedulerCache, and should stay as such.
type nodeTree struct {
	tree      map[string]*nodeArray // a map from zone (region-zone) to an array of nodes in the zone.
	zones     []string              // a list of all the zones in the tree (keys)
	zoneIndex int
	numNodes  int
}

// nodeArray is a struct that has nodes that are in a zone.
// We use a slice (as opposed to a set/map) to store the nodes because iterating over the nodes is
// a lot more frequent than searching them by name.
type nodeArray struct {
	nodes     []string
	lastIndex int
}

func (na *nodeArray) next() (nodeName string, exhausted bool) {
	if len(na.nodes) == 0 {
		logger.Errorf("The nodeArray is empty. It should have been deleted from NodeTree.")
		return "", false
	}
	if na.lastIndex >= len(na.nodes) {
		return "", true
	}
	nodeName = na.nodes[na.lastIndex]
	na.lastIndex++
	return nodeName, false
}

// newNodeTree creates a NodeTree from nodes.
func newNodeTree(nodes []*types.SiteNode) *nodeTree {
	nt := &nodeTree{
		tree: make(map[string]*nodeArray),
	}
	for _, n := range nodes {
		nt.addNode(n)
	}
	return nt
}

// addNode adds a node and its corresponding zone to the tree. If the zone already exists, the node
// is added to the array of nodes in that zone.
func (nt *nodeTree) addNode(n *types.SiteNode) {
	zone := utils.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for _, nodeName := range na.nodes {
			if nodeName == n.SiteID {
				return
			}
		}
		na.nodes = append(na.nodes, n.SiteID)
	} else {
		nt.zones = append(nt.zones, zone)
		nt.tree[zone] = &nodeArray{nodes: []string{n.SiteID}, lastIndex: 0}
	}
	logger.Infof("Added node %q in group %q to NodeTree", n.SiteID, zone)
	nt.numNodes++
}

// removeNode removes a node from the NodeTree.
func (nt *nodeTree) removeNode(n *types.SiteNode) error {
	zone := utils.GetZoneKey(n)
	if na, ok := nt.tree[zone]; ok {
		for i, nodeName := range na.nodes {
			if nodeName == n.SiteID {
				na.nodes = append(na.nodes[:i], na.nodes[i+1:]...)
				if len(na.nodes) == 0 {
					nt.removeZone(zone)
				}
				logger.Infof("Removed node %q in group %q from NodeTree", n.SiteID, zone)
				nt.numNodes--
				return nil
			}
		}
	}
	logger.Errorf("Node %q in group %q was not found", n.SiteID, zone)
	return fmt.Errorf("node %q in group %q was not found", n.SiteID, zone)
}

// removeZone removes a zone from tree.
// This function must be called while writer locks are hold.
func (nt *nodeTree) removeZone(zone string) {
	delete(nt.tree, zone)
	for i, z := range nt.zones {
		if z == zone {
			nt.zones = append(nt.zones[:i], nt.zones[i+1:]...)
			return
		}
	}
}

// updateNode updates a node in the NodeTree.
func (nt *nodeTree) updateNode(old, new *types.SiteNode) {
	var oldZone string
	if old != nil {
		oldZone = utils.GetZoneKey(old)
	}
	newZone := utils.GetZoneKey(new)
	// If the zone ID of the node has not changed, we don't need to do anything. Name of the node
	// cannot be changed in an update.
	if oldZone == newZone {
		return
	}
	err := nt.removeNode(old) // No error checking. We ignore whether the old node exists or not.
	if err != nil {
		logger.Errorf("removeNode failed! err: %s", err)
	}
	nt.addNode(new)
}

func (nt *nodeTree) resetExhausted() {
	for _, na := range nt.tree {
		na.lastIndex = 0
	}
	nt.zoneIndex = 0
}

// next returns the name of the next node. NodeTree iterates over zones and in each zone iterates
// over nodes in a round robin fashion.
func (nt *nodeTree) next() string {
	if len(nt.zones) == 0 {
		return ""
	}
	numExhaustedZones := 0
	for {
		if nt.zoneIndex >= len(nt.zones) {
			nt.zoneIndex = 0
		}
		zone := nt.zones[nt.zoneIndex]
		nt.zoneIndex++
		// We do not check the exhausted zones before calling next() on the zone. This ensures
		// that if more nodes are added to a zone after it is exhausted, we iterate over the new nodes.
		nodeName, exhausted := nt.tree[zone].next()
		if exhausted {
			numExhaustedZones++
			if numExhaustedZones >= len(nt.zones) { // all zones are exhausted. we should reset.
				nt.resetExhausted()
			}
		} else {
			return nodeName
		}
	}
}
