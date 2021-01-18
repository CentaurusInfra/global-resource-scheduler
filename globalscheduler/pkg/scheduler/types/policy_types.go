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

package types

// LabelsPresence holds the parameters that are used to configure the corresponding predicate
// in scheduler policy configuration.
type LabelsPresence struct {
	// The list of labels that identify site "groups"
	// All of the labels should be either present (or absent) for the site to be considered a fit for hosting the pod
	Labels []string
	// The boolean flag that indicates whether the labels should be present or absent from the site
	Presence bool
}

// PredicateArgument represents the arguments to configure predicate functions in scheduler policy configuration.
// Only one of its members may be specified
type PredicateArgument struct {
	// The predicate that checks whether a particular site has a certain label
	// defined or not, regardless of value
	LabelsPresence *LabelsPresence
}

// PriorityPolicy describes a struct of a priority policy.
type PriorityPolicy struct {
	// Identifier of the priority policy
	// For a custom priority, the name can be user-defined
	// For the Kubernetes provided priority functions, the name is the identifier of the pre-defined priority function
	Name string
	// The numeric multiplier for the site scores that the priority function generates
	// The weight should be a positive integer
	Weight int64
	// Holds the parameters to configure the given priority function
	Argument *PredicateArgument
}

// Policy describes a struct of a policy resource in api.
type Policy struct {
	// Holds the information to configure the fit predicate functions.
	// If unspecified, the default predicate functions will be applied.
	// If empty list, all predicates (except the mandatory ones) will be
	// bypassed.
	Predicates []PredicatePolicy
	// Holds the information to configure the priority functions.
	// If unspecified, the default priority functions will be applied.
	// If empty list, all priority functions will be bypassed.
	Priorities []PriorityPolicy
}

// PredicatePolicy describes a struct of a predicate policy.
type PredicatePolicy struct {
	// Identifier of the predicate policy
	// For a custom predicate, the name can be user-defined
	// For the Kubernetes provided predicates, the name is the identifier of the pre-defined predicate
	Name string
	// Holds the parameters to configure the given predicate
	Argument *PredicateArgument
}
