load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["geoLocationDistribute.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/controllers/util/geoLocationDistribute",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/apis/cluster/v1:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/v1:go_default_library",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
