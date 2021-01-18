load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "globaltracer.go",
        "gocontext.go",
        "noop.go",
        "propagation.go",
        "span.go",
        "tracer.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/opentracing/opentracing-go",
    importpath = "github.com/opentracing/opentracing-go",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/opentracing/opentracing-go/log:go_default_library"],
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
        "//vendor/github.com/opentracing/opentracing-go/ext:all-srcs",
        "//vendor/github.com/opentracing/opentracing-go/log:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
