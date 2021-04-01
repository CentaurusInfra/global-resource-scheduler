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
	archaiuscore "github.com/go-chassis/go-archaius/core"
	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

var conf archaius.ConfigurationFactory
var FlavorMap map[string]*typed.RegionFlavor

type Listener struct {
	Key string
}

func Init() {
	err := archaius.Init()
	if err != nil {
		klog.Error("[ERROR]archaius init failed\n")
		os.Exit(1)
	}
	conf = archaius.GetConfigFactory()
	confLocation := utils.GetConfigDirectory()
	klog.Infof("Config file directory: %v\n", confLocation)
	confFilePath := filepath.Join(confLocation, "conf.yaml")
	err = archaius.AddFile(confFilePath)
	if err != nil {
		klog.Errorf("failed to AddFile %s, error: %+v\n", confFilePath, err.Error())
		os.Exit(1)
	}
	klog.Infof("Config file %s\n", confFilePath)

	regionFilePath := filepath.Join(confLocation, "region.yaml")
	err = archaius.AddFile(regionFilePath)
	if err != nil {
		klog.Errorf("failed to AddFile %s, error: %+v\n", regionFilePath, err.Error())
		os.Exit(1)
	}
	klog.Infof("Region Config File %s\n", regionFilePath)

	flavorConfigFilePath := filepath.Join(confLocation, "flavor_config.yaml")
	err = archaius.AddFile(flavorConfigFilePath)
	if err != nil {
		klog.Errorf("failed to AddFile %s, error: %+v\n", flavorConfigFilePath, err.Error())
		os.Exit(1)
	}
	klog.Infof("Region Config File %s\n", flavorConfigFilePath)
	ReadFlavorConf()
	archaius.RegisterListener(&Listener{}, "flavor_number")
}

func (e Listener) Event(event *archaiuscore.Event) {
	klog.Infof("Flavor Config Event - event key: %v", event)
	if event.EventType == archaiuscore.Create || event.EventType == archaiuscore.Update {
		ReadFlavorConf()
	}
}

// String convert key to string
func String(key string) string {
	value, err := conf.GetValue(key).ToString()
	if err != nil {
		klog.Errorf("get %s from config failed! err: %s\n", key, err.Error())
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
		klog.Infof("get %s from config failed! err: %s\n", key, err.Error())
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

func ReadFlavorConf() map[string]*typed.RegionFlavor {
	if FlavorMap == nil {
		FlavorMap = make(map[string]*typed.RegionFlavor)
	}
	flavor_number := archaius.GetInt("flavor_number", 0)
	var id, name, vcpus, vdisk string
	var ram int64
	for i := 0; i < flavor_number; i++ {
		flavor_index := fmt.Sprintf("flavors.flavor_%v", i)
		id = archaius.GetString(flavor_index+".id", "42")
		name = archaius.GetString(flavor_index+".name", "m1.nano")
		vcpus = archaius.GetString(flavor_index+".vcpus", "1")
		ram = int64(archaius.GetInt(flavor_index+".ram", 128))
		vdisk = archaius.GetString(flavor_index+".vdisk", "0")
		flavor := &typed.RegionFlavor{
			RegionFlavorID: id,
			Region:         "",
			Flavor: typed.Flavor{
				ID:    id,
				Name:  name,
				Vcpus: vcpus,
				Ram:   ram,
				Disk:  vdisk,
			},
		}
		FlavorMap[id] = flavor
	}
	klog.Infof("FlavorMap: %v", FlavorMap)
	return FlavorMap
}
