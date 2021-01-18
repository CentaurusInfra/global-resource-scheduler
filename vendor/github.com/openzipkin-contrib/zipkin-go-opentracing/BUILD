load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "collector.go",
        "collector-http.go",
        "collector-kafka.go",
        "collector-scribe.go",
        "context.go",
        "debug.go",
        "event.go",
        "log-materializers.go",
        "logger.go",
        "observer.go",
        "propagation.go",
        "propagation_ot.go",
        "raw.go",
        "recorder.go",
        "sample.go",
        "span.go",
        "tracer.go",
        "util.go",
        "zipkin-endpoint.go",
        "zipkin-recorder.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/openzipkin-contrib/zipkin-go-opentracing",
    importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/Shopify/sarama:go_default_library",
        "//vendor/github.com/apache/thrift/lib/go/thrift:go_default_library",
        "//vendor/github.com/go-logfmt/logfmt:go_default_library",
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
        "//vendor/github.com/opentracing-contrib/go-observer:go_default_library",
        "//vendor/github.com/opentracing/opentracing-go:go_default_library",
        "//vendor/github.com/opentracing/opentracing-go/ext:go_default_library",
        "//vendor/github.com/opentracing/opentracing-go/log:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/flag:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/scribe:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/types:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/wire:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/flag:all-srcs",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/scribe:all-srcs",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore:all-srcs",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/types:all-srcs",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/wire:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
