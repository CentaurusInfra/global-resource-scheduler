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

package app

var (
	// Version is ieg service current version
	Version   string
	commitID  string
	buildDate string
	goVersion string

	// Info produce environment info
	Info = ProdInfo{
		Version:   Version,
		CommitID:  commitID,
		BuildDate: buildDate,
		GoVersion: goVersion,
	}
)

// ProdInfo is the Produce Environment basic information
type ProdInfo struct {
	Version   string `json:"version" description:"-"`
	CommitID  string `json:"commit_id" description:"-"`
	BuildDate string `json:"build_date" description:"-"`
	GoVersion string `json:"go_version" description:"-"`
}
