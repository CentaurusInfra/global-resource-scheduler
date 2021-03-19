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

package router

import (
	"strings"

	_ "k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/config"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/service"

	"github.com/emicklei/go-restful"
	_ "k8s.io/klog"
)

// Route define four basic info of the northbound interface
// which is used to invoke web service.
type Route struct {
	Name      string
	Method    string
	Pattern   string
	RouteFunc restful.RouteFunction
}

// Register routes
func Register() {
	ws := new(restful.WebService)
	ws.Path("/").Consumes("*/*").Produces("*/*")
	registerRoute(ws)
	restful.Add(ws)
}

func registerRoute(ws *restful.WebService) {
	for _, route := range routes {
		ws.Route(ws.Method(route.Method).Path(route.Pattern).To(route.RouteFunc).Operation(route.Name))
	}
}

// Routes is an array of Route
type Routes []Route

var routes = Routes{
	Route{
		"Allocations",
		strings.ToUpper("Post"),
		"/allocations",
		service.Allocations,
	},
	Route{
		"PushSnapshot",
		strings.ToUpper("Patch"),
		"/globalscheduler/v1/regionresources/{regionname}",
		service.PushSnapshot,
	},
	Route{
		"PushUpdatedResources",
		strings.ToUpper("Post"),
		"/regionresources",
		service.PushRegionResources,
	},
}
