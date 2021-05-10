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

package controllers

import (
	"fmt"
	"time"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateCRD(cs clientset.Interface, crd *apiextv1beta1.CustomResourceDefinition) error {
	if err := createCRD(cs, crd); err != nil {
		return err
	}
	return waitForCRD(cs, crd)
}

func createCRD(cs clientset.Interface, crd *apiextv1beta1.CustomResourceDefinition) error {
	_, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if apierrors.IsAlreadyExists(err) {
		currentCRD, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		crd.ResourceVersion = currentCRD.ResourceVersion
		_, err = cs.ApiextensionsV1beta1().CustomResourceDefinitions().Update(crd)
		return err
	} else if err != nil {
		return err
	}

	return nil
}

func waitForCRD(cs clientset.Interface, crd *apiextv1beta1.CustomResourceDefinition) error {
	operation := func() (bool, error) {
		manifest, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for _, cond := range manifest.Status.Conditions {
			switch cond.Type {
			case apiextv1beta1.Established:
				if cond.Status == apiextv1beta1.ConditionTrue {
					return true, nil
				}
			case apiextv1beta1.NamesAccepted:
				if cond.Status == apiextv1beta1.ConditionFalse {
					return false, fmt.Errorf("name %s conflicts", crd.Name)
				}
			}
		}

		return false, fmt.Errorf("%s not established with the final conditions %v", crd.Name, manifest.Status.Conditions)
	}
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		return operation()
	})

	if err != nil {
		deleteErr := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(crd.Name, nil)
		if deleteErr != nil {
			return errors.NewAggregate([]error{err, deleteErr})
		}
		return err
	}
	return err
}

func CreateAllocationCRD() *apiextv1beta1.CustomResourceDefinition {
	return &apiextv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "allocations.globalscheduler.com",
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group:   "globalscheduler.com",
			Version: "v1",
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Kind:   "Allocation",
				Plural: "allocations",
			},
			Validation: &apiextv1beta1.CustomResourceValidation{
				OpenAPIV3Schema: &apiextv1beta1.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextv1beta1.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextv1beta1.JSONSchemaProps{
								"resource_group": {
									Type: "object",
									Properties: map[string]apiextv1beta1.JSONSchemaProps{
										"name": {Type: "string"},
										"resources": {
											Type: "array",
											Items: &apiextv1beta1.JSONSchemaPropsOrArray{
												Schema: &apiextv1beta1.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextv1beta1.JSONSchemaProps{
														"name":          {Type: "string"},
														"resource_type": {Type: "string"},
														"flavors": {
															Type: "array",
															Items: &apiextv1beta1.JSONSchemaPropsOrArray{
																Schema: &apiextv1beta1.JSONSchemaProps{
																	Type: "object",
																	Properties: map[string]apiextv1beta1.JSONSchemaProps{
																		"flavor_id": {Type: "string"},
																		"spot": {
																			Type: "object",
																			Properties: map[string]apiextv1beta1.JSONSchemaProps{
																				"max_price":           {Type: "string"},
																				"spot_duration_hours": {Type: "integer"},
																				"spot_duration_count": {Type: "integer"},
																				"interruption_policy": {Type: "string"},
																			},
																		},
																	},
																},
															},
														},
														"storage": {
															Type: "object",
															Properties: map[string]apiextv1beta1.JSONSchemaProps{
																"sata": {Type: "integer"},
																"sas":  {Type: "integer"},
																"ssd":  {Type: "integer"},
															},
														},
														"need_eip": {Type: "boolean"},
													},
												},
											},
										},
									},
								},
								"selector": {
									Type: "object",
									Properties: map[string]apiextv1beta1.JSONSchemaProps{
										"geo_location": {
											Type: "object",
											Properties: map[string]apiextv1beta1.JSONSchemaProps{
												"city":     {Type: "string"},
												"province": {Type: "string"},
												"area":     {Type: "string"},
												"country":  {Type: "string"},
											},
										},
										"regions": {
											Type: "array",
											Items: &apiextv1beta1.JSONSchemaPropsOrArray{
												Schema: &apiextv1beta1.JSONSchemaProps{
													Type: "object",
													Properties: map[string]apiextv1beta1.JSONSchemaProps{
														"region": {Type: "string"},
														"availability_zone": {
															Type: "array",
															Items: &apiextv1beta1.JSONSchemaPropsOrArray{
																Schema: &apiextv1beta1.JSONSchemaProps{Type: "string"},
															},
														},
													},
												},
											},
										},
										"operator": {Type: "string"},
										"strategy": {
											Type: "object",
											Properties: map[string]apiextv1beta1.JSONSchemaProps{
												"location_strategy": {Type: "string"},
											},
										},
									},
								},
								"replicas": {Type: "integer"},
							},
						},
						"status": {
							Type: "object",
							Properties: map[string]apiextv1beta1.JSONSchemaProps{
								"phase":            {Type: "string"},
								"distributor_name": {Type: "string"},
								"dispatcher_name":  {Type: "string"},
								"scheduler_name":   {Type: "string"},
								"cluster_names": {
									Type: "array",
									Items: &apiextv1beta1.JSONSchemaPropsOrArray{
										Schema: &apiextv1beta1.JSONSchemaProps{Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			AdditionalPrinterColumns: []apiextv1beta1.CustomResourceColumnDefinition{
				{
					Name:     "hashkey",
					Type:     "string",
					JSONPath: ".metadata.hashKey",
				},
				{
					Name:     "city",
					Type:     "string",
					JSONPath: ".spec.selector.geo_location.city",
				},
				{
					Name:     "province",
					Type:     "string",
					JSONPath: ".spec.selector.geo_location.province",
				},
				{
					Name:     "area",
					Type:     "string",
					JSONPath: ".spec.selector.geo_location.area",
				},
				{
					Name:     "country",
					Type:     "string",
					JSONPath: ".spec.selector.geo_location.country",
				},
				{
					Name:     "status",
					Type:     "string",
					JSONPath: ".status.phase",
				},
				{
					Name:     "distributor",
					Type:     "string",
					JSONPath: ".status.distributor_name",
				},
				{
					Name:     "scheduler",
					Type:     "string",
					JSONPath: ".status.scheduler_name",
				},
				{
					Name:     "dispatcher",
					Type:     "string",
					JSONPath: ".status.dispatcher_name",
				},
			},
		},
	}
}
