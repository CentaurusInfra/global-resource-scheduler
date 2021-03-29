/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package config

import (
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chassis/go-archaius"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

var conf archaius.ConfigurationFactory

//var GlobalConf *Config //http server

func Init() {
	err := archaius.Init()
	if err != nil {
		fmt.Printf("[ERROR]archaius init failed\n")
		os.Exit(1)
	}
	conf = archaius.GetConfigFactory()

	confLocation := utils.GetConfigDirectory()
	confFilePath := filepath.Join(confLocation, "conf.yaml")
	err = archaius.AddFile(confFilePath)
	if err != nil {
		fmt.Printf("failed to AddFile %s, error: %+v\n", confFilePath, err.Error())
		os.Exit(1)
	}

	regionFilePath := filepath.Join(confLocation, "region.yaml")
	err = archaius.AddFile(regionFilePath)
	if err != nil {
		fmt.Printf("failed to AddFile %s, error: %+v\n", regionFilePath, err.Error())
		os.Exit(1)
	}

	//http server
	//GlobalConf = readConf()
}

// String convert key to string
func String(key string) string {
	value, err := conf.GetValue(key).ToString()
	if err != nil {
		fmt.Printf("get %s from config failed! err: %s\n", key, err.Error())
		return ""
	}
	return value
}

// StringMap convert string to map[string]interface{}
func StringMap(key string) map[string]interface{} {

	var retMap = map[string]interface{}{}
	configs := archaius.GetConfigs()
	for index, value := range configs {
		splitStr := strings.Split(index, ".")
		if len(splitStr) == 2 && splitStr[0] == key {
			retMap[splitStr[1]] = value
		}
	}

	return retMap
}

// ArrayMapString convert string to []map[string]string
func ArrayMapString(key string) []map[string]string {
	var m = []map[string]string{}
	value, err := conf.GetValue(key).ToSlice()
	if err != nil {
		fmt.Printf("get %s from config failed! err: %s\n", key, err.Error())
		return nil
	}

	for _, ms := range value {
		var mapStringItem = map[string]string{}
		switch v := ms.(type) {
		case yaml.MapSlice:
			for _, item := range v {
				mapStringItem[cast.ToString(item.Key)] = cast.ToString(item.Value)
			}
			m = append(m, mapStringItem)
		case map[string]interface{}:
			for key, value := range v {
				mapStringItem[cast.ToString(key)] = cast.ToString(value)
			}
			m = append(m, mapStringItem)
		default:
			klog.Errorf("unable to cast %#v of type %T to yaml.MapItem", v, v)
		}
	}
	return m
}

// DefaultString convert key with defaultValue string to string
func DefaultString(key string, defaultValue string) string {
	value, err := conf.GetValue(key).ToString()
	if err != nil {
		return defaultValue
	}

	return value
}

// DefaultBool convert key with defaultValue bool to string
func DefaultBool(key string, defaultVal bool) bool {
	value, err := conf.GetValue(key).ToBool()
	if err != nil {
		return defaultVal
	}

	return value
}

// DefaultInt convert key with defaultValue int to string
func DefaultInt(key string, defaultValue int) int {
	value, err := conf.GetValue(key).ToInt()
	if err != nil {
		return defaultValue
	}

	return value
}

// DefaultInt64 convert key with defaultValue int64 to string
func DefaultInt64(key string, defaultValue int64) int64 {
	value, err := conf.GetValue(key).ToInt64()
	if err != nil {
		return defaultValue
	}

	return value
}

// DefaultFloat64 convert key with defaultValue float64 to string
func DefaultFloat64(key string, defaultValue float64) float64 {
	value, err := conf.GetValue(key).ToFloat64()
	if err != nil {
		return defaultValue
	}

	return value
}

// InitPolicyFromFile initialize policy from file
func InitPolicyFromFile(policyFile string, policy *types.Policy) error {
	// Use a policy serialized in a file.
	_, err := os.Stat(policyFile)
	if err != nil {
		return fmt.Errorf("missing policy config file %s", policyFile)
	}
	data, err := ioutil.ReadFile(policyFile)
	if err != nil {
		return fmt.Errorf("couldn't read policy config: %v", err)
	}

	err = yaml.Unmarshal(data, policy)
	if err != nil {
		return fmt.Errorf("invalid policy: %v", err)
	}
	return nil
}

//http server
/*type Config struct {
	// HTTP Server
	HttpAddr   string
	HttpPort   int
	APIVersion string
}

func readConf() *Config {
	return &Config{
		HttpAddr:   DefaultString("HttpAddr", "0.0.0.0"),
		HttpPort:   DefaultInt("HttpPort", 8661),
		APIVersion: DefaultString("APIVersion", "v1"),
	}
} */

func ReadFlavorConf() map[string]*typed.RegionFlavor {
	FlavorMap := make(map[string]*typed.RegionFlavor)
	flavor42 := &typed.RegionFlavor{
		RegionFlavorID: "42",
		Region:         "",
		Flavor: typed.Flavor{
			ID: "42",

			// Specifies the name of the ECS specifications.
			Name: "42",

			// Specifies the number of CPU cores in the ECS specifications.
			Vcpus: "1",

			// Specifies the memory size (MB) in the ECS specifications.
			Ram: 128,

			// Specifies the system disk size in the ECS specifications.
			// The value 0 indicates that the disk size is not limited.
			Disk: "0",

			/*// Specifies shortcut links for ECS flavors.
			Links []Link `json:"links"`

			// Specifies extended ECS specifications.
			OsExtraSpecs OsExtraSpecs `json:"os_extra_specs"`

			// Reserved
			Swap string `json:"swap"`

			// Reserved
			FlvEphemeral int64 `json:"OS-FLV-EXT-DATA:ephemeral"`

			// Reserved
			FlvDisabled bool `json:"OS-FLV-DISABLED:disabled"`

			// Reserved
			RxtxFactor int64 `json:"rxtx_factor"`

			// Reserved
			RxtxQuota string `json:"rxtx_quota"`

			// Reserved
			RxtxCap string `json:"rxtx_cap"`

			// Reserved
			AccessIsPublic bool `json:"os-flavor-access:is_public"`*/
		},
	}
	flavor1 := &typed.RegionFlavor{
		RegionFlavorID: "1",
		Region:         "",
		Flavor: typed.Flavor{
			ID: "1",

			// Specifies the name of the ECS specifications.
			Name: "m1.tiny",

			// Specifies the number of CPU cores in the ECS specifications.
			Vcpus: "1",

			// Specifies the memory size (MB) in the ECS specifications.
			Ram: 512,

			// Specifies the system disk size in the ECS specifications.
			// The value 0 indicates that the disk size is not limited.
			Disk: "1",

			/*// Specifies shortcut links for ECS flavors.
			Links []Link `json:"links"`

			// Specifies extended ECS specifications.
			OsExtraSpecs OsExtraSpecs `json:"os_extra_specs"`

			// Reserved
			Swap string `json:"swap"`

			// Reserved
			FlvEphemeral int64 `json:"OS-FLV-EXT-DATA:ephemeral"`

			// Reserved
			FlvDisabled bool `json:"OS-FLV-DISABLED:disabled"`

			// Reserved
			RxtxFactor int64 `json:"rxtx_factor"`

			// Reserved
			RxtxQuota string `json:"rxtx_quota"`

			// Reserved
			RxtxCap string `json:"rxtx_cap"`

			// Reserved
			AccessIsPublic bool `json:"os-flavor-access:is_public"`*/
		},
	}
	flavor2 := &typed.RegionFlavor{
		RegionFlavorID: "2",
		Region:         "",
		Flavor: typed.Flavor{
			ID: "2",

			// Specifies the name of the ECS specifications.
			Name: "m1.small",

			// Specifies the number of CPU cores in the ECS specifications.
			Vcpus: "1",

			// Specifies the memory size (MB) in the ECS specifications.
			Ram: 2048,

			// Specifies the system disk size in the ECS specifications.
			// The value 0 indicates that the disk size is not limited.
			Disk: "20",

			/*// Specifies shortcut links for ECS flavors.
			Links []Link `json:"links"`

			// Specifies extended ECS specifications.
			OsExtraSpecs OsExtraSpecs `json:"os_extra_specs"`

			// Reserved
			Swap string `json:"swap"`

			// Reserved
			FlvEphemeral int64 `json:"OS-FLV-EXT-DATA:ephemeral"`

			// Reserved
			FlvDisabled bool `json:"OS-FLV-DISABLED:disabled"`

			// Reserved
			RxtxFactor int64 `json:"rxtx_factor"`

			// Reserved
			RxtxQuota string `json:"rxtx_quota"`

			// Reserved
			RxtxCap string `json:"rxtx_cap"`

			// Reserved
			AccessIsPublic bool `json:"os-flavor-access:is_public"`*/
		},
	}
	FlavorMap["42"] = flavor42
	FlavorMap["1"] = flavor1
	FlavorMap["2"] = flavor2
	return FlavorMap
}
