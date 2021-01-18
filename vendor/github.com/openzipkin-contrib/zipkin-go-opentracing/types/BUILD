load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["traceid.go"],
    importmap = "k8s.io/kubernetes/vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/types",
    importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing/types",
    visibility = ["//visibility:public"],
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
