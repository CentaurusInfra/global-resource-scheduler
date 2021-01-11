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

package scheduler

import (
	"fmt"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"time"
)

func (sc *SchedulerController) doesSchedulerCRDExist() (bool, error) {
	schedulerCrd, err := sc.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get("schedulers.globalscheduler.com", metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	// Check whether the CRD is accepted.
	for _, condition := range schedulerCrd.Status.Conditions {
		if condition.Type == apiextensions.Established &&
			condition.Status == apiextensions.ConditionTrue {
			return true, nil
		}
	}

	return false, fmt.Errorf("CRD is not accepted")
}

func (sc *SchedulerController) waitSchedulerCRDAccepted() error {
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		return sc.doesSchedulerCRDExist()
	})

	return err
}

func (sc *SchedulerController) CreateSchedulerCRD() error {
	if result, _ := sc.doesSchedulerCRDExist(); result {
		return nil
	}

	schedulerCrd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "schedulers.globalscheduler.com",
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   "globalscheduler.com",
			Version: "v1",
			Names: apiextensions.CustomResourceDefinitionNames{
				Kind:   "Scheduler",
				Plural: "schedulers",
			},
			Validation: &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextensions.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextensions.JSONSchemaProps{
								"location": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"city":     {Type: "string"},
										"province": {Type: "string"},
										"area":     {Type: "string"},
										"country":  {Type: "string"},
									},
								},
								"tag": {Type: "string"},
								"cluster": {
									Type: "array",
									Items: &apiextensions.JSONSchemaPropsOrArray{
										Schema: &apiextensions.JSONSchemaProps{Type: "string"},
									},
								},
								"union": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"geolocation": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextensions.JSONSchemaProps{
														"city":     {Type: "string"},
														"province": {Type: "string"},
														"area":     {Type: "string"},
														"country":  {Type: "string"},
													},
												},
											},
										},
										"region": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextensions.JSONSchemaProps{
														"region":           {Type: "string"},
														"availabilityzone": {Type: "string"},
													},
												},
											},
										},
										"operator": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextensions.JSONSchemaProps{
														"operator": {Type: "string"},
													},
												},
											},
										},
										"flavors": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextensions.JSONSchemaProps{
														"flavorid":      {Type: "string"},
														"totalcapacity": {Type: "integer"},
													},
												},
											},
										},
										"storage": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextensions.JSONSchemaProps{
														"typeid":          {Type: "string"},
														"storagecapacity": {Type: "integer"},
													},
												},
											},
										},
										"eipcapacity": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{Type: "integer"},
											},
										},
										"cpucapacity": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{Type: "integer"},
											},
										},
										"memcapacity": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{Type: "integer"},
											},
										},
										"serverprice": {
											Type: "array",
											Items: &apiextensions.JSONSchemaPropsOrArray{
												Schema: &apiextensions.JSONSchemaProps{Type: "integer"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := sc.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(schedulerCrd)

	if err != nil {
		klog.Fatalf(err.Error())
	}

	return sc.waitSchedulerCRDAccepted()
}
