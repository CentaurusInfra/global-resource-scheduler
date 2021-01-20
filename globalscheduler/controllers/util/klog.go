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
	"flag"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

func InitKlog(namespace, name, logfile, loglevel string) {
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("log_file", logfile+"."+namespace+"_"+name)
	flag.Set("v", loglevel)
	go wait.Forever(klog.Flush, 5*time.Second)
}

func FlushKlog() {
	klog.Flush()
}
