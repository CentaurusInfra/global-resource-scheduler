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
