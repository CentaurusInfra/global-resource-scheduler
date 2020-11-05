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
