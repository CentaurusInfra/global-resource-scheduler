package logger

import (
	"context"
	"fmt"
	"strings"

	l4g "k8s.io/klog"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger/trace"
)

// remove \r or \n
func processString(format string, args ...interface{}) string {
	printStr := fmt.Sprintf(format, args...)

	printStr = strings.Replace(printStr, "\r", "", -1)
	printStr = strings.Replace(printStr, "\n", "", -1)

	return printStr
}

// Infof print info level log
func Infof(format string, args ...interface{}) {
	var traceID = ""

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Info(processString("[req-%s] "+format, newArgs...))
}

// Info print info level log with context
func Info(ctx context.Context, format string, args ...interface{}) {
	var traceID = ""
	if ctx != nil {
		traceID = trace.GetTraceID(ctx)
	}

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Info(processString("[req-%s] "+format, newArgs...))
}

// Warnf print info warn level log
func Warnf(format string, args ...interface{}) {
	var traceID = ""

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Warningf(processString("[req-%s] "+format, newArgs...))
}

// Warn print warn level log with context
func Warn(ctx context.Context, format string, args ...interface{}) {
	var traceID = ""
	if ctx != nil {
		traceID = trace.GetTraceID(ctx)
	}

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Warningf(processString("[req-%s] "+format, newArgs...))
}

// Debugf print info warn level log
func Debugf(format string, args ...interface{}) {
	var traceID = ""

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Infof(processString("[req-%s] "+format, newArgs...))
}

// Debug print debug level log with context
func Debug(ctx context.Context, format string, args ...interface{}) {
	var traceID = ""
	if ctx != nil {
		traceID = trace.GetTraceID(ctx)
	}

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Infof(processString("[req-%s] "+format, newArgs...))
}

// Errorf print info error level log
func Errorf(format string, args ...interface{}) {
	var traceID = ""

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Error(processString("[req-%s] "+format, newArgs...))
}

// Error print error level log with context
func Error(ctx context.Context, format string, args ...interface{}) {
	var traceID = ""
	if ctx != nil {
		traceID = trace.GetTraceID(ctx)
	}

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Error(processString("[req-%s] "+format, newArgs...))
}

// Fatalf print info fatal level log
func Fatalf(format string, args ...interface{}) {
	var traceID = ""

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Fatal(processString("[req-%s] "+format, newArgs...))
}

// Fatal print fatal level log with context
func Fatal(ctx context.Context, format string, args ...interface{}) {
	var traceID = ""
	if ctx != nil {
		traceID = trace.GetTraceID(ctx)
	}

	newArgs := append([]interface{}{traceID}, args...)
	l4g.Fatalf("[req-%s] "+format, newArgs...)
}
