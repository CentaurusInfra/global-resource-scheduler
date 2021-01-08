/*
Copyright 2019 The Kubernetes Authors.
Copyright 2020 Authors of Arktos - file modified.

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

package trace

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/opentracing/opentracing-go"
	"github.com/openzipkin-contrib/zipkin-go-opentracing"
)

func CatchException() {
	err := recover()
	GetExceptionMsg(err)
}

func GetExceptionMsg(err interface{}) string {
	if err != nil {
		errMsg := fmt.Sprintf("Exception: %+v, call stack: %v", err, string(debug.Stack()))
		fmt.Println(errMsg)
		return errMsg
	}
	return ""
}

//GetTraceID return trace ID
func GetTraceID(ctx context.Context) string {
	defer CatchException()
	span := opentracing.SpanFromContext(ctx)

	if nil == span {
		return ""
	}

	sctx := span.Context()
	if nil == sctx {
		return ""
	}

	var traceID string
	switch sctx.(type) {
	case zipkintracer.SpanContext:
		zinkinSc := sctx.(zipkintracer.SpanContext)
		traceID = zinkinSc.TraceID.ToHex()
	default:
		traceID = ""
	}

	return traceID
}
