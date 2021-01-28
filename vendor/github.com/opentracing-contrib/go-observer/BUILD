load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["observer.go"],
    importmap = "k8s.io/kubernetes/vendor/github.com/opentracing-contrib/go-observer",
    importpath = "github.com/opentracing-contrib/go-observer",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/opentracing/opentracing-go:go_default_library"],
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
