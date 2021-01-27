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

package constants

const (
	// HeaderParameterToken is the request's header title key X-Auth-Token
	HeaderParameterToken = "X-Auth-Token"

	// HeaderContentType is the request's header title key Content-Type
	HeaderContentType = "Content-Type"

	// ContextDomainID is the request's context param domain_id
	ContextDomainID = "domain_id"

	// ContextToken is the request's context param token
	ContextToken = "token"

	// ContextRegions is the request's context param regions
	ContextRegions = "regions"

	//ContextNeedAdmin is admin
	ContextNeedAdmin = "needAdmin"

	//TimeFormat is time str format
	TimeFormat = "2006-01-02 15:04:05"

	//StrategyDiscrete
	StrategyDiscrete = "discrete"

	//StrategyCentralize
	StrategyCentralize = "centralize"

	//StrategyRegionAlone
	StrategyRegionAlone = "alone"

	/***************config file key parameter****************/

	ConfPolicyFile = "scheduler_policy_file"

	ConfQosWeight = "qos_weight"

	//SiteStatusNormal normal
	SiteStatusNormal = "Normal"

	SiteStatusOffline = "Offline"

	SiteStatusSellout = "Sellout"

	//NameMaxLen is max name len
	NameMaxLen = 64

	// NamePattern uses to check name
	NamePattern = "^[a-zA-Z0-9_.\u4e00-\u9fa5-]*$"

	SiteExclusiveDomains = "exclusive_domains"

	InterruptionPolicyDelay = "delay"

	//ConfVolumeTypeInterval volumetype_interval, default 600s
	ConfVolumeTypeInterval = "volumetype_interval"

	//ConfSiteInfoInterval site_interval, default 600s
	ConfSiteInfoInterval = "siteinfo_interval"

	//ConfFlavorInterval flavor_interval, default 600s
	ConfFlavorInterval = "flavor_interval"

	//ConfEipPoolInterval eippool_interval, default 60s
	ConfEipPoolInterval = "eippool_interval"

	//ConfVolumePoolInterval volumepool_interval, default 60s
	ConfVolumePoolInterval = "volumepool_interval"

	//ConfCommonHypervisorInterval common_hypervisor_interval, default 24h，86400s
	ConfCommonHypervisorInterval = "common_hypervisor_interval"

	//Default resource collector api url
	DefaultResourceCollectorAPIURL = "127.0.0.1:8663"

	//ResourceCollector snapshot url
	ResourceCollecotrSnapshotURL = "/v1/snapshot"
)
