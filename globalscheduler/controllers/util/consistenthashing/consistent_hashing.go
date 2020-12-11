/*
Copyright 2020 Authors of Arktos.

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

// Consistent hashing is a special kind of hashing such that when a hash table is resized,
// only K/n keys need to be remapped on average,
// where K is the number of keys, and  n is the number of slots.

package consistenthashing

import (
	"errors"
	"hash/fnv"
	"sort"
	"strconv"
	"sync"
)

const VIRTUAL_NODE_NUMBER = 131072

type uints []uint32

// Len returns the length of the uints array.
func (x uints) Len() int { return len(x) }

// Less returns true if element i is less than element j.
func (x uints) Less(i, j int) bool { return x[i] < x[j] }

// Swap exchanges elements i and j.
func (x uints) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

type ConsistentHash struct {
	HashCircle           map[uint32]string // store hash key & value pair
	SortedHashes         uints             // sorted hash key
	NumberOfVirtualNodes int               // virtual nodes number
	Members              []string
	Results              map[string][]string
	sync.RWMutex         // read & write lock
}

// New creates a new Consistent object with a default setting of 131,072 replicas for each entry.
//
// To change the number of replicas, set NumberOfVirtualNodes before adding entries.
func New() *ConsistentHash {
	ch := new(ConsistentHash)
	ch.NumberOfVirtualNodes = VIRTUALNODENUMBER
	ch.HashCircle = make(map[uint32]string)
	ch.Members = []string{}
	ch.Results = make(map[string][]string)
	return ch
}

// 32-byte Fowler-Noll-Vo hash algorithm
func (ch *ConsistentHash) fnv32Hash(key string) uint32 {
	new32Hash := fnv.New32()
	new32Hash.Write([]byte(key))
	return new32Hash.Sum32()
}

func (ch *ConsistentHash) generateKey(elt string, idx int) string {
	return elt + "#" + strconv.Itoa(idx)
}

func (ch *ConsistentHash) updateSortedHashes() {
	// Improve space complexity without reallocation
	hashes := ch.SortedHashes[:0]

	// Reallocate if sortedHashed array is too big
	if cap(ch.SortedHashes)/(ch.NumberOfVirtualNodes*4) > len(ch.HashCircle) {
		hashes = nil
	}
	for k := range ch.HashCircle {
		hashes = append(hashes, k)
	}
	sort.Sort(hashes)
	ch.SortedHashes = hashes
}

func (ch *ConsistentHash) Add(elt string) {
	ch.Lock()
	defer ch.Unlock()

	ch.add(elt)
}
func (ch *ConsistentHash) add(elt string) {
	// add virtual nodes
	for i := 0; i < ch.NumberOfVirtualNodes; i++ {
		ch.HashCircle[ch.fnv32Hash(ch.generateKey(elt, i))] = elt
	}

	// sort hash key
	ch.updateSortedHashes()

	if len(ch.Members) > 0 {
		err := ch.rebalance()
		if err != nil {
			return
		}
	}
}

func (ch *ConsistentHash) Remove(elt string) {
	ch.Lock()
	defer ch.Unlock()

	ch.remove(elt)
}
func (ch *ConsistentHash) remove(elt string) {
	for i := 0; i < ch.NumberOfVirtualNodes; i++ {
		delete(ch.HashCircle, ch.fnv32Hash(ch.generateKey(elt, i)))
	}

	// sort hash key
	ch.updateSortedHashes()

	if len(ch.Members) > 0 {
		err := ch.rebalance()
		if err != nil {
			return
		}
	}
}

func (ch *ConsistentHash) Insert(name string) (string, error) {
	ch.RLock()
	defer ch.RUnlock()

	if len(ch.HashCircle) == 0 {
		return "", errors.New("empty hash circle")
	}
	ch.Members = append(ch.Members, name)

	key := ch.fnv32Hash(name)
	// find the first index of hash value that is greater than key
	idx := ch.search(key)

	ch.Results[ch.HashCircle[ch.SortedHashes[idx]]] = append(ch.Results[ch.HashCircle[ch.SortedHashes[idx]]], name)

	return ch.HashCircle[ch.SortedHashes[idx]], nil
}

func (ch *ConsistentHash) GetIdList(name string) []string {
	if ch.Results != nil {
		return ch.Results[name]
	}

	return nil
}

func (ch *ConsistentHash) search(key uint32) int {
	f := func(x int) bool {
		return ch.SortedHashes[x] > key
	}

	sortedHashedLength := len(ch.SortedHashes)
	idx := sort.Search(sortedHashedLength, f)
	if idx < sortedHashedLength {
		return idx
	} else {
		return 0
	}
}

func (ch *ConsistentHash) rebalance() error {
	ch.Results = make(map[string][]string)
	for _, v := range ch.Members {
		_, err := ch.Insert(v)
		if err != nil {
			return err
		}
	}
	return nil
}
