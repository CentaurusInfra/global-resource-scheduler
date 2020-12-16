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

package consistenthashing

import (
	"sort"
	"strconv"
	"testing"
)

func checkNum(num, expected int, t *testing.T) {
	if num != expected {
		t.Errorf("got %d, expected %d", num, expected)
	}
}

func TestNew(t *testing.T) {
	x := New()
	if x == nil {
		t.Errorf("expected obj")
	}
	checkNum(x.NumberOfVirtualNodes, VIRTUAL_NODE_NUMBER, t)
}

func TestAdd(t *testing.T) {
	x := New()
	input := []string{"node-1"}
	x.Add(input)
	checkNum(len(x.HashCircle), VIRTUAL_NODE_NUMBER, t)
	checkNum(len(x.SortedHashes), VIRTUAL_NODE_NUMBER, t)
	if sort.IsSorted(x.SortedHashes) == false {
		t.Errorf("expected sorted hashes to be sorted")
	}
	input = []string{"node-2"}
	x.Add(input)
	checkNum(len(x.HashCircle), VIRTUAL_NODE_NUMBER*2, t)
	checkNum(len(x.SortedHashes), VIRTUAL_NODE_NUMBER*2, t)
	if sort.IsSorted(x.SortedHashes) == false {
		t.Errorf("expected sorted hashes to be sorted")
	}
}

func TestRemove(t *testing.T) {
	x := New()
	input := []string{"node-1"}
	x.Add(input)
	x.Remove("node-1")
	checkNum(len(x.HashCircle), 0, t)
	checkNum(len(x.SortedHashes), 0, t)
}

func TestRemoveNonExisting(t *testing.T) {
	x := New()
	input := []string{"node-1"}
	x.Add(input)
	x.Remove("node-3")
	checkNum(len(x.HashCircle), VIRTUAL_NODE_NUMBER, t)
}

func TestInsertEmpty(t *testing.T) {
	x := New()
	input := []string{"node-1"}
	err := x.Insert(input)
	if err == nil {
		t.Errorf("expected error")
	}
	checkNum(len(x.Members), 0, t)
}

func TestDelete(t *testing.T) {
	x := New()
	input := []string{"add-1"}
	x.Add(input)
	input = []string{"insert-1"}
	err := x.Insert(input)
	if err != nil {
		t.Errorf("expected no error")
	}
	checkNum(len(x.Members), 1, t)

	x.Delete("insert-1")
	checkNum(len(x.Members), 0, t)
}

func TestMemberNum(t *testing.T) {
	x := New()
	input := []string{"node-1"}
	x.Add(input)
	inputNum := 10
	var res []string
	for i := 0; i < inputNum; i++ {
		s := strconv.Itoa(i)
		res = append(res, s)
	}

	err := x.Insert(res)
	if err != nil {
		t.Errorf("expected no error")
	}
	checkNum(len(x.Members), inputNum, t)
}
