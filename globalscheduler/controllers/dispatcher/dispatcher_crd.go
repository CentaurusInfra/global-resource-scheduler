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

package dispatcher

import (
	"fmt"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

func (dc *DispatcherController) doesDispatcherCRDExist() (bool, error) {
	dispatcherCrd, err := dc.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get("dispatchers.globalscheduler.com", metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	// Check whether the CRD is accepted.
	for _, condition := range dispatcherCrd.Status.Conditions {
		if condition.Type == apiextensions.Established &&
			condition.Status == apiextensions.ConditionTrue {
			return true, nil
		}
	}

	return false, fmt.Errorf("CRD is not accepted")
}

func (dc *DispatcherController) waitDispatcherCRDAccepted() error {
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		return dc.doesDispatcherCRDExist()
	})

	return err
}

func (dc *DispatcherController) CreateDispatcherCRD() error {
	if result, _ := dc.doesDispatcherCRDExist(); result {
		return nil
	}

	dispatcherCrd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dispatchers.globalscheduler.com",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   "globalscheduler.com",
			Version: "v1",
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:   "Dispatcher",
				Plural: "dispatchers",
			},
			Validation: &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextensions.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextensions.JSONSchemaProps{
								"clusterRange": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"start": {Type: "string"},
										"end":   {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "ClusterRange",
					Type:     "string",
					JSONPath: ".spec.clusterRange",
				},
			},
		},
	}

	_, err := dc.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(dispatcherCrd)

	if err != nil {
		klog.Fatalf(err.Error())
	}

	return dc.waitDispatcherCRDAccepted()
}
