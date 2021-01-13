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

package cache

// ListerInterface is tracked to handle late initialization of the controller
type ListerInterface interface {
	List(options interface{}) ([]interface{}, error)
}

// ListFunc knows how to list resources
type ListFunc func(options interface{}) ([]interface{}, error)

// Lister is defined to invoke ListFunc
type Lister struct {
	ListFunc ListFunc
}

// List a set of apiServer resources
func (lw *Lister) List(options interface{}) ([]interface{}, error) {
	return lw.ListFunc(options)
}
