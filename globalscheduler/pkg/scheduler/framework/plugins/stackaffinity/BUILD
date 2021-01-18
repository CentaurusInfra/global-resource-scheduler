load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["stackaffinity.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins/stackaffinity",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/interfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/labels:go_default_library",
        "//globalscheduler/pkg/scheduler/sitecacheinfo:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
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
