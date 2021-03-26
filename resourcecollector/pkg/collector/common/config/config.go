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
	"os"
	"path/filepath"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"

	"github.com/go-chassis/go-archaius"
)

var conf archaius.ConfigurationFactory
var GlobalConf *Config

func Init() {
	err := archaius.Init()
	if err != nil {
		fmt.Printf("[ERROR]archaius init failed\n")
		os.Exit(1)
	}
	conf = archaius.GetConfigFactory()

	confLocation := utils.GetConfigDirectory()
	confFilePath := filepath.Join(confLocation, "resource_collector_conf.yaml")
	err = archaius.AddFile(confFilePath)
	if err != nil {
		fmt.Printf("failed to AddFile %s, error: %+v\n", confFilePath, err.Error())
		os.Exit(1)
	}

	GlobalConf = readConf()
}

type Config struct {
	// RPC Server
	RpcPort string

	// HTTP Server
	HttpAddr   string
	HttpPort   int
	APIVersion string

	// ClusterController Info
	ClusterControllerIP   string
	ClusterControllerPort string

	// OpenStack
	OpenStackUsername        string
	OpenStackPassword        string
	OpenStackDomainID        string
	OpenStackScopProjectName string
	OpenStackScopDomainID    string

	// Informer interval period
	FlavorInterval       int
	SiteResourceInterval int
	VolumePoolInterval   int
	VolumeTypeInterval   int
	EipPoolInterval      int

	// When the maximum number of unreachable requests is reached,
	// send the node unreachable status to the ClusterController
	MaxUnreachableNum int

	// Refresh OpenStack token interval, unit: minutes
	RefreshOpenStackTokenInterval int
}

func readConf() *Config {
	return &Config{
		RpcPort: DefaultString("RpcPort", "50052"),

		HttpAddr:   DefaultString("HttpAddr", "0.0.0.0"),
		HttpPort:   DefaultInt("HttpPort", 8663),
		APIVersion: DefaultString("APIVersion", "v1"),

		ClusterControllerIP:   DefaultString("ClusterControllerIP", "127.0.0.1"),
		ClusterControllerPort: DefaultString("ClusterControllerPort", "50053"),

		OpenStackUsername:        DefaultString("OpenStackUsername", "admin"),
		OpenStackPassword:        DefaultString("OpenStackPassword", "secret"),
		OpenStackDomainID:        DefaultString("OpenStackDomainID", "default"),
		OpenStackScopProjectName: DefaultString("OpenStackScopProjectName", "admin"),
		OpenStackScopDomainID:    DefaultString("OpenStackScopDomainID", "default"),

		FlavorInterval:       DefaultInt("FlavorInterval", 60),
		SiteResourceInterval: DefaultInt("SiteResourceInterval", 60),
		VolumePoolInterval:   DefaultInt("VolumePoolInterval", 60),
		VolumeTypeInterval:   DefaultInt("VolumeTypeInterval", 60),
		EipPoolInterval:      DefaultInt("EipPoolInterval", 60),

		MaxUnreachableNum: DefaultInt("MaxUnreachableNum", 5),

		RefreshOpenStackTokenInterval: DefaultInt("RefreshOpenStackTokenInterval", 30),
	}
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
