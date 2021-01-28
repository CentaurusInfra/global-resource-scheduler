load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "GoUnusedProtection__.go",
        "zipkinCore.go",
        "zipkinCore-consts.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore",
    importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing/thrift/gen-go/zipkincore",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/apache/thrift/lib/go/thrift:go_default_library"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
