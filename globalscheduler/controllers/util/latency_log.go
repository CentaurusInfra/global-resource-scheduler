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
	"k8s.io/klog"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func CheckTime(podName string, moduleName string, funcName string, start int) {
	checkTime := time.Now().UTC()
	duration := checkTime.UnixNano() / 1000000
	latency := int(duration)
	klog.Infof("\n LatencyLog %s %s %s %v %v \n", podName, moduleName, funcName, start, latency)

	// filename := flag.String("file", "./cluster.yaml", "cluster file name")
	// f, err := os.Create(*filename)
	// check(err)
	// defer f.Close()
	// _, err = f.WriteString("  homescheduler: scheduler0\n")
}
