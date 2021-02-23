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
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/globalscheduler/controllers"
)

func (dc *DispatcherController) CreateDispatcherCRD() error {

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
					Name:     "cluster_name_start",
					Type:     "string",
					JSONPath: ".spec.clusterRange.start",
				},
				{
					Name:     "cluster_name_end",
					Type:     "string",
					JSONPath: ".spec.clusterRange.end",
				},
			},
		},
	}
	return controllers.CreateCRD(dc.apiextensionsclientset, dispatcherCrd)
}
