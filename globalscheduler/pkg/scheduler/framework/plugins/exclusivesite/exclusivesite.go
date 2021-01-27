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

package exclusivesite

import (
	"context"
	"strings"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/constants"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/types"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/sets"
)

// Name is the name of the plugin used in the plugin registry and configurations.
const Name = "ExclusiveSite"

// PrioritySort is a plugin that implements Priority based sorting.
type ExclusiveSite struct{}

var _ interfaces.FilterPlugin = &ExclusiveSite{}

// Name returns name of the plugin.
func (pl *ExclusiveSite) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
func (pl *ExclusiveSite) Filter(ctx context.Context, cycleState *interfaces.CycleState, stack *types.Stack,
	siteCacheInfo *sitecacheinfo.SiteCacheInfo) *interfaces.Status {
	domainID := utils.GetStrFromCtx(ctx, constants.ContextDomainID)

	siteAttrs := siteCacheInfo.GetSite().SiteAttribute

	for _, siteAttr := range siteAttrs {
		if siteAttr.Key == constants.SiteExclusiveDomains {
			whiteList := siteAttr.Value
			allowDomains := strings.Split(whiteList, ",")
			allowDomainsSet := sets.NewString(allowDomains...)
			if allowDomainsSet.Has(domainID) {
				logger.Debug(ctx, "Site(%s-%s) belong to domianID(%s)", siteCacheInfo.GetSite().SiteID,
					siteCacheInfo.GetSite().Region, domainID)
				return interfaces.NewStatus(interfaces.Unschedulable, "site is exclusive site.")
			}

			break
		}
	}

	return nil
}

// New initializes a new plugin and returns it.
func New(handle interfaces.FrameworkHandle) (interfaces.Plugin, error) {
	return &ExclusiveSite{}, nil
}
