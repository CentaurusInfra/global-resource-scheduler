load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client/informers/externalversions/internalinterfaces:go_default_library",
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
        "//globalscheduler/pkg/apis/cluster/client/informers/externalversions/cluster/v1:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
