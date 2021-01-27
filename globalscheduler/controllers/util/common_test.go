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

package util

import (
	"math"
	"reflect"
	"testing"
)

type Testcase struct {
	Size           int
	ExpectedResult [][]int64
}

func TestEvenlyDivide0(t *testing.T) {
	testcases := []Testcase{
		{Size: 0, ExpectedResult: make([][]int64, 0)},
		{Size: 1, ExpectedResult: [][]int64{{0, 0}}},
		{Size: 2, ExpectedResult: [][]int64{{0, 0}}},
		{Size: 3, ExpectedResult: [][]int64{{0, 0}}},
		{Size: 4, ExpectedResult: [][]int64{{0, 0}}},
		{Size: 5, ExpectedResult: [][]int64{{0, 0}}},
		{Size: 6, ExpectedResult: [][]int64{{0, 0}}},
		{Size: 7, ExpectedResult: [][]int64{{0, 0}}},
	}
	for _, testcase := range testcases {
		res := EvenlyDivide(testcase.Size, 0)
		if testcase.Size > 0 && !reflect.DeepEqual(res, testcase.ExpectedResult) {
			t.Errorf("The test result %v is not expected as %v", res, testcase.ExpectedResult)
		} else if testcase.Size == 0 {
			if len(res) != 0 {
				t.Errorf("The test result %v is not empty as expected", res)
			}
		}
	}
}

func TestEvenlyDivide1(t *testing.T) {
	testcases := []Testcase{
		{Size: 0, ExpectedResult: make([][]int64, 0)},
		{Size: 1, ExpectedResult: [][]int64{{0, 1}}},
		{Size: 2, ExpectedResult: [][]int64{{0, 0}, {1, 1}}},
		{Size: 3, ExpectedResult: [][]int64{{0, 0}, {1, 1}}},
		{Size: 4, ExpectedResult: [][]int64{{0, 0}, {1, 1}}},
		{Size: 5, ExpectedResult: [][]int64{{0, 0}, {1, 1}}},
		{Size: 6, ExpectedResult: [][]int64{{0, 0}, {1, 1}}},
		{Size: 7, ExpectedResult: [][]int64{{0, 0}, {1, 1}}},
	}
	for _, testcase := range testcases {
		res := EvenlyDivide(testcase.Size, 1)
		if testcase.Size > 0 && !reflect.DeepEqual(res, testcase.ExpectedResult) {
			t.Errorf("The test result %v is not expected as %v", res, testcase.ExpectedResult)
		} else if testcase.Size == 0 {
			if len(res) != 0 {
				t.Errorf("The test result %v is not empty as expected", res)
			}
		}
	}
}

func TestEvenlyDivide2(t *testing.T) {
	testcases := []Testcase{
		{Size: 0, ExpectedResult: make([][]int64, 0)},
		{Size: 1, ExpectedResult: [][]int64{{0, 2}}},
		{Size: 2, ExpectedResult: [][]int64{{0, 1}, {2, 2}}},
		{Size: 3, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}}},
		{Size: 4, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}}},
		{Size: 5, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}}},
		{Size: 6, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}}},
		{Size: 7, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}}},
	}
	for _, testcase := range testcases {
		res := EvenlyDivide(testcase.Size, 2)
		if testcase.Size > 0 && !reflect.DeepEqual(res, testcase.ExpectedResult) {
			t.Errorf("The test result %v is not expected as %v", res, testcase.ExpectedResult)
		} else if testcase.Size == 0 {
			if len(res) != 0 {
				t.Errorf("The test result %v is not empty as expected", res)
			}
		}
	}
}

func TestEvenlyDivide3(t *testing.T) {
	testcases := []Testcase{
		{Size: 0, ExpectedResult: make([][]int64, 0)},
		{Size: 1, ExpectedResult: [][]int64{{0, 3}}},
		{Size: 2, ExpectedResult: [][]int64{{0, 1}, {2, 3}}},
		{Size: 3, ExpectedResult: [][]int64{{0, 1}, {2, 2}, {3, 3}}},
		{Size: 4, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}, {3, 3}}},
		{Size: 5, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}, {3, 3}}},
		{Size: 6, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}, {3, 3}}},
		{Size: 7, ExpectedResult: [][]int64{{0, 0}, {1, 1}, {2, 2}, {3, 3}}},
	}
	for _, testcase := range testcases {
		res := EvenlyDivide(testcase.Size, 3)
		if testcase.Size > 0 && !reflect.DeepEqual(res, testcase.ExpectedResult) {
			t.Errorf("The test result %v is not expected as %v", res, testcase.ExpectedResult)
		} else if testcase.Size == 0 {
			if len(res) != 0 {
				t.Errorf("The test result %v is not empty as expected", res)
			}
		}
	}
}

func TestEvenlyDivide100(t *testing.T) {
	testcases := []Testcase{
		{Size: 0, ExpectedResult: make([][]int64, 0)},
		{Size: 1, ExpectedResult: [][]int64{{0, 100}}},
		{Size: 2, ExpectedResult: [][]int64{{0, 50}, {51, 100}}},
		{Size: 3, ExpectedResult: [][]int64{{0, 33}, {34, 67}, {68, 100}}},
		{Size: 4, ExpectedResult: [][]int64{{0, 25}, {26, 50}, {51, 75}, {76, 100}}},
		{Size: 5, ExpectedResult: [][]int64{{0, 20}, {21, 40}, {41, 60}, {61, 80}, {81, 100}}},
		{Size: 6, ExpectedResult: [][]int64{{0, 16}, {17, 33}, {34, 50}, {51, 67}, {68, 84}, {85, 100}}},
		{Size: 7, ExpectedResult: [][]int64{{0, 14}, {15, 29}, {30, 44}, {45, 58}, {59, 72}, {73, 86}, {87, 100}}},
	}
	for _, testcase := range testcases {
		res := EvenlyDivide(testcase.Size, 100)
		if testcase.Size > 0 && !reflect.DeepEqual(res, testcase.ExpectedResult) {
			t.Errorf("The test result %v is not expected as %v", res, testcase.ExpectedResult)
		} else if testcase.Size == 0 {
			if len(res) != 0 {
				t.Errorf("The test result %v is not empty as expected", res)
			}
		}
	}
}

func TestEvenlyDivideInt64(t *testing.T) {
	testcases := make([]Testcase, 7)
	testcases[0] = Testcase{Size: 0, ExpectedResult: make([][]int64, 0)}
	testcases[1] = Testcase{Size: 1, ExpectedResult: [][]int64{{0, math.MaxInt64}}}
	testcases[2] = Testcase{Size: 2, ExpectedResult: [][]int64{{0, math.MaxInt64 / 2}, {math.MaxInt64/2 + 1, math.MaxInt64}}}
	chunk := math.MaxInt64 / int64(3)
	testcases[3] = Testcase{Size: 3, ExpectedResult: [][]int64{{0, chunk}, {chunk + 1, 2*chunk + 1}, {2*chunk + 2, math.MaxInt64}}}
	chunk = math.MaxInt64 / int64(4)
	testcases[4] = Testcase{Size: 4, ExpectedResult: [][]int64{{0, chunk}, {chunk + 1, 2*chunk + 1}, {2*chunk + 2, 3*chunk + 2}, {3*chunk + 3, math.MaxInt64}}}
	chunk = math.MaxInt64 / int64(5)
	testcases[5] = Testcase{Size: 5, ExpectedResult: [][]int64{{0, chunk}, {chunk + 1, 2*chunk + 1}, {2*chunk + 2, 3*chunk + 2}, {3*chunk + 3, 4*chunk + 2}, {4*chunk + 3, math.MaxInt64}}}
	chunk = math.MaxInt64 / int64(6)
	testcases[6] = Testcase{Size: 6, ExpectedResult: [][]int64{{0, chunk}, {chunk + 1, 2*chunk + 1}, {2*chunk + 2, 3*chunk + 1}, {3*chunk + 2, 4*chunk + 1}, {4*chunk + 2, 5*chunk + 1}, {5*chunk + 2, math.MaxInt64}}}
	for _, testcase := range testcases {
		res := EvenlyDivideInt64(testcase.Size)
		if testcase.Size > 0 && !reflect.DeepEqual(res, testcase.ExpectedResult) {
			t.Errorf("The test result %v is not expected as %v", res, testcase.ExpectedResult)
		} else if testcase.Size == 0 {
			if len(res) != 0 {
				t.Errorf("The test result %v is not empty as expected", res)
			}
		}
	}
}
