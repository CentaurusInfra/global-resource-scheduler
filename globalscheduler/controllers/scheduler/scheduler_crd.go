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
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
	"time"
)

func (sc *SchedulerController) doesClusterCRDExist() (bool, error) {
	clusterCrd, err := sc.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(clusterv1.Name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	// Check whether the CRD is accepted.
	for _, condition := range clusterCrd.Status.Conditions {
		if condition.Type == apiextensions.Established &&
			condition.Status == apiextensions.ConditionTrue {
			return true, nil
		}
	}

	return false, fmt.Errorf("CRD is not accepted")
}

func (sc *SchedulerController) waitClusterCRDAccepted() error {
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		return sc.doesClusterCRDExist()
	})

	return err
}

// CreateCRD creates a custom resource definition, Cluster.
func (sc *SchedulerController) CreateClusterCRD() error {
	if result, _ := sc.doesClusterCRDExist(); result {
		return nil
	}
	clusterCrd := &apiextensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterv1.Name,
		},
		Spec: apiextensions.CustomResourceDefinitionSpec{
			Group:   clusterv1.GroupName,
			Version: clusterv1.Version,
			Scope:   apiextensions.NamespaceScoped,
			Names: apiextensions.CustomResourceDefinitionNames{
				Plural:     clusterv1.Plural,
				Singular:   clusterv1.Singluar,
				Kind:       clusterv1.Kind,
				ShortNames: []string{clusterv1.ShortName},
			},
			Validation: &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
					Type: "object",
					Properties: map[string]apiextensions.JSONSchemaProps{
						"spec": {
							Type: "object",
							Properties: map[string]apiextensions.JSONSchemaProps{
								"ipaddress": {Type: "string"},
								"geolocation": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"city":     {Type: "string"},
										"province": {Type: "string"},
										"area":     {Type: "string"},
										"country":  {Type: "string"},
									},
								},
								"region": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"region":           {Type: "string"},
										"availabilityzone": {Type: "string"},
									},
								},
								"operator": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"operator": {Type: "string"},
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
								"eipcapacity":   {Type: "integer"},
								"cpucapacity":   {Type: "integer"},
								"memcapacity":   {Type: "integer"},
								"serverprice":   {Type: "integer"},
								"homescheduler": {Type: "string"},
							},
						},
					},
				},
			},
			AdditionalPrinterColumns: []apiextensions.CustomResourceColumnDefinition{
				{
					Name:     "ipaddress",
					Type:     "string",
					JSONPath: ".spec.ipaddress",
				},
				{
					Name:     "region",
					Type:     "string",
					JSONPath: ".spec.region",
				},
				{
					Name:     "eipcapacity",
					Type:     "integer",
					JSONPath: ".spec.eipcapacity",
				},
				{
					Name:     "cpucapacity",
					Type:     "integer",
					JSONPath: ".spec.cpucapacity",
				},
				{
					Name:     "memcapacity",
					Type:     "integer",
					JSONPath: ".spec.memcapacity",
				},
				{
					Name:     "serverprice",
					Type:     "integer",
					JSONPath: ".spec.serverprice",
				},
				{
					Name:     "homescheduler",
					Type:     "string",
					JSONPath: ".spec.homescheduler",
				},
				{
					Name:     "age",
					Type:     "date",
					JSONPath: ".metadata.createTimestamp",
				},
			},
		},
	}

	_, err := sc.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(clusterCrd)

	if err != nil {
		klog.Fatalf(err.Error())
	}
	return sc.waitClusterCRDAccepted()
}

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
