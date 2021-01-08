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

package utils

// ContentKey is defined to solve the goLint warning: should not use basic type string as ContentKey in context.WithValue
type ContentKey string

const (

	// ContextRegions is the request's context param regions
	ContextRegions ContentKey = "regions"

	//ContextNeedAdmin is admin
	ContextNeedAdmin ContentKey = "needAdmin"

	// GlobalRegion global
	GlobalRegion = "global"

	// RegionID region_id key
	RegionID = "region_id"

	// TimeFormat is time str format
	TimeFormat = "2006-01-02 15:04:05"
)

// DefaultRegion default region
var DefaultRegion = ""
