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

import "math"

func RemoveCluster(clusters []string, clusterName string) []string {
	var idx int
	for i, v := range clusters {
		if v == clusterName {
			idx = i
			break
		}
	}
	clusters[idx], clusters[len(clusters)-1] = clusters[len(clusters)-1], clusters[idx]
	return clusters[:len(clusters)-1]
}

func EvenlyDivideInt64(size int) [][]int64 {
	return EvenlyDivide(size, math.MaxInt64)
}

func EvenlyDivide(size int, upper int64) [][]int64 {
	res := make([][]int64, 0)
	if size <= 0 {
		return res
	}

	// hash function can only get uint32, uint64
	// k8s code base does not deal with uint32 properly
	// uint64 > MaxInt64 will have issue in converter. Need to map to 0 - maxInt64
	var start int64 = 0
	var end int64 = 0
	chunk := upper / int64(size)
	mod := upper % int64(size)
	i := 0
	for i < size {
		end = start + chunk - 1
		if int64(i) <= mod {
			end += 1
		}
		if start > upper || end > upper {
			return res
		}
		resItem := make([]int64, 2)
		resItem[0] = start
		resItem[1] = end
		res = append(res, resItem)
		i = i + 1
		start = end + 1
	}
	return res
}
