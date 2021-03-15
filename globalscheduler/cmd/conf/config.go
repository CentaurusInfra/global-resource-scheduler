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
	"github.com/spf13/viper"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog"
	"sync"
)

var lock = &sync.Mutex{}
var configInstance *Config

type Config struct {
	Qps QpsConfig
}

func New(configFile string) *Config {
	lock.Lock()
	defer lock.Unlock()
	if configInstance == nil {
		viper.SetConfigFile(configFile)
		var conf Config

		if err := viper.ReadInConfig(); err != nil {
			klog.Fatalf("Failed to read config file %s with the error %v", configFile, err)
			return nil
		}
		err := viper.Unmarshal(&conf)
		if err != nil {
			klog.Fatalf("Failed to read config file %s with the error %v", configFile, err)
			return nil
		}
		configInstance = &conf
	}
	return configInstance
}

func AddQPSFlags(config *restclient.Config, configFile string) {
	New(configFile)
	kubeConfigs := config.GetAllConfigs()
	for _, kubeConfig := range kubeConfigs {
		kubeConfig.ContentType = configInstance.Qps.ContentType
		kubeConfig.QPS = configInstance.Qps.Qps
		kubeConfig.Burst = configInstance.Qps.Burst
	}
}
