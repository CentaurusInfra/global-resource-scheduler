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
	"net"
	"strings"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/config"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/service"

	"github.com/emicklei/go-restful"
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
	apiVersion := "/" + config.DefaultString("component_version", "v1")

	ws := new(restful.WebService)
	ws.Path(apiVersion).Consumes(restful.MIME_JSON, restful.MIME_XML).Produces(restful.MIME_JSON, restful.MIME_XML)
	registerRoute(ws)
	restful.Add(ws)
}

func getRealIP(req *restful.Request) string {
	xRealIP := req.Request.Header.Get("X-Real-ID")
	xForwardedFor := req.Request.Header.Get("X-Forwarded-For")

	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		if address != "" {
			return address
		}
	}

	if xRealIP != "" {
		return xRealIP
	}

	ip, _, err := net.SplitHostPort(req.Request.RemoteAddr)
	if err != nil {
		ip = req.Request.RemoteAddr
	}
	if ip != "127.0.0.1" {
		return ip
	}

	return "-"
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
}
