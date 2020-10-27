load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/apis/clusters/informers/externalversions/clusters",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/globalscheduler/pkg/apis/clusters/informers/externalversions/clusters/v1:go_default_library",
        "//vendor/globalscheduler/pkg/apis/clusters/informers/externalversions/internalinterfaces:go_default_library",
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
        "//globalscheduler/pkg/apis/clusters/informers/externalversions/clusters/v1:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
