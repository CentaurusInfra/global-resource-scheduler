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

package client

// Interface is db row PROJECT_IDaccess to a entrance for all southbound interface.
// It consists of all kinds of southbound interface.
type Interface interface {
}

// ClientSet provides access to a client set for all southbound interface.
// It consists of all kind of southbound client.
type ClientSet struct {
}
