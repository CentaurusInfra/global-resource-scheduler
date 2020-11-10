package controller

import (
	"fmt"
	"time"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	clusterv1 "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/v1"
)

func (c *ClusterController) doesCRDExist() (bool, error) {
	crd, err := c.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(vpcv1.Name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	// Check whether the CRD is accepted.
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensions.Established &&
			condition.Status == apiextensions.ConditionTrue {
			return true, nil
		}
	}

	return false, fmt.Errorf("CRD is not accepted")
}

func (c *ClusterController) waitCRDAccepted() error {
	err := wait.Poll(1*time.Second, 10*time.Second, func() (bool, error) {
		return c.doesCRDExist()
	})

	return err
}

// CreateCRD creates a custom resource definition, Vpc.
func (c *ClusterController) CreateCRD() error {
	if result, _ := c.doesCRDExist(); result {
		return nil
	}

	crd := &apiextensions.CustomResourceDefinition{
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
									Required: []string{"area", "country"},
								},
								"region": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"region":           {Type: "string"},
										"availabilityzone": {Type: "[]string"},
									},
								},
								"operator": {
									Type: "object",
									Properties: map[string]apiextensions.JSONSchemaProps{
										"operator": {Type: "string"},
									},
									Required: []string{"operator"},
								},
								"flavors": {
									Type: "array",
									Items: &apiextensions.JSONSchemaPropsOrArray{
										Schema: &apiextensions.JSONSchemaProps{
											Type: "object",
											Properties: map[string]apiextensions.JSONSchemaProps{
												"flavorid":      {Type: "string"},
												"totalcapacity": {Type: "int64"},
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
												"storagecapacity": {Type: "int64"},
											},
										},
									},
								},
								"eipcapacity":   {Type: "int64"},
								"cpucapacity":   {Type: "int64"},
								"memcapacity":   {Type: "int64"},
								"serverprice":   {Type: "int64"},
								"homescheduler": {Type: "string"},
							},
							Required: []string{"ipaddress", "region", "storage"},
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
					Type:     "int64",
					JSONPath: ".spec.eipcapacity",
				},
				{
					Name:     "cpucapacity",
					Type:     "int64",
					JSONPath: ".spec.cpucapacity",
				},
				{
					Name:     "memcapacity",
					Type:     "int64",
					JSONPath: ".spec.memcapacity",
				},
				{
					Name:     "serverprice",
					Type:     "int64",
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

	_, err := c.apiextensionsclientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)

	if err != nil {
		klog.Fatalf(err.Error())
	}
	return c.waitCRDAccepted()
}
