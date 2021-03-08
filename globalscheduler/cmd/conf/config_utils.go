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

package conf

import (
	"flag"
	restclient "k8s.io/client-go/rest"
)

func AddQPSFlags(config *restclient.Config) {
	qps := flag.Int("qps", 20, "The maximum QPS to the master from this client (default 1e+06)")
	burst := flag.Int("burst", 40, "The maximum burst for throttle (default 1000000)")
	contentType := flag.String("content_type", "application/vnd.kubernetes.protobuf", "The content type")
	kubeConfigs := config.GetAllConfigs()
	for _, kubeConfig := range kubeConfigs {
		kubeConfig.ContentType = *contentType
		kubeConfig.QPS = float32(*qps)
		kubeConfig.Burst = int(*burst)
	}
}
