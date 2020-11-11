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

package service

import (
	"context"
	"fmt"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/typed"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

//CheckIDOrName Check whether the ID and Name are valid.
func CheckIDOrName(ctx context.Context, allocation types.Allocation) error {
	if utils.IsUUIDValid(allocation.ID) != nil {
		logger.Error(ctx, "allocation id(%#v) is invalid.", allocation.ID)
		return fmt.Errorf("allocation id is invalid")
	}

	// check stack label info
	for key, value := range allocation.Stack.Labels {
		err := utils.CheckTitleLengthAndRegexValid(key, 1, constants.NameMaxLen, constants.NamePattern)
		if err != nil {
			logger.Error(ctx, "allocation stack Lable key(%#v) is invalid, err: %s", key, err)
			return fmt.Errorf("allocation stack Lable key is invalid")
		}

		err = utils.CheckTitleLengthAndRegexValid(value, 1, constants.NameMaxLen, constants.NamePattern)
		if err != nil {
			logger.Error(ctx, "allocation stack Lable value(%#v) is invalid, err: %s", value, err)
			return fmt.Errorf("allocation stack Lable value is invalid")
		}
	}

	err := utils.CheckTitleLengthAndRegexValid(allocation.Stack.Name, 1, constants.NameMaxLen, constants.NamePattern)
	if err != nil {
		logger.Error(ctx, "allocation stack Name(%#v) is invalid, err: %s", allocation.Stack.Name, err)
		return fmt.Errorf("allocation stack name is invalid")
	}

	for _, oneServer := range allocation.Stack.Resources {
		err := utils.CheckTitleLengthAndRegexValid(oneServer.Name, 1, constants.NameMaxLen, constants.NamePattern)
		if err != nil {
			logger.Error(ctx, "server name(%#v) is invalid. err: %s", oneServer.Name, err)
			return err
		}
	}

	return nil
}

// CheckStrategy Check whether the policy field is valid.
func CheckStrategy(ctx context.Context, allocation types.Allocation) error {

	switch allocation.Selector.Strategy.LocationStrategy {
	case constants.StrategyCentralize, constants.StrategyDiscrete, "":
		return nil
	}

	return fmt.Errorf("Strategy(%#v) not support", allocation.Selector.Strategy.LocationStrategy)
}

// Check whether the flavor is valid.
func CheckFlavor(ctx context.Context, allocation types.Allocation) error {
	if len(allocation.Stack.Resources) <= 0 {
		return nil
	}

	flavors := informers.InformerFac.GetInformer(informers.FLAVOR).GetStore().List()

	for _, oneResource := range allocation.Stack.Resources {
		for _, flv := range oneResource.Flavors {
			var isFind = false
			for _, fla := range flavors {
				regionFlv, ok := fla.(typed.RegionFlavor)
				if !ok {
					continue
				}

				if regionFlv.ID == flv.FlavorID {
					isFind = true
					break
				}
			}

			if !isFind {
				msg := fmt.Sprintf("flavor(%#v) does not support!", flv)
				logger.Error(ctx, msg)
				return fmt.Errorf(msg)
			}
		}
	}

	return nil
}

func checkVolumeType(voType string, volumeTypes []interface{}) bool {
	if volumeTypes == nil {
		return false
	}

	for _, one := range volumeTypes {
		if typeObj, ok := one.(typed.RegionVolumeType); ok {
			if typeObj.Name == voType {
				return true
			}
		}
	}

	return false
}

// CheckStorage
func CheckStorage(ctx context.Context, allocation types.Allocation) error {
	if len(allocation.Stack.Resources) <= 0 {
		return nil
	}

	volumeTypes := informers.InformerFac.GetInformer(informers.VOLUMETYPE).GetStore().List()
	for _, oneServer := range allocation.Stack.Resources {
		for key, value := range oneServer.Storage {
			if !checkVolumeType(key, volumeTypes) {
				msg := fmt.Sprintf("volumeType(%#v) not support!", key)
				logger.Error(ctx, msg)
				return fmt.Errorf(msg)
			}

			if value <= 0 {
				msg := fmt.Sprintf("value(%#v) of volumeType(%#v) is invalid!", value, key)
				logger.Error(ctx, msg)
				return fmt.Errorf(msg)
			}
		}
	}

	return nil
}
