load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["sitecache_info.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/sitecacheinfo",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/client/informers:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
        "//globalscheduler/pkg/scheduler/common/config:go_default_library",
        "//globalscheduler/pkg/scheduler/common/constants:go_default_library",
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//globalscheduler/pkg/scheduler/utils:go_default_library",
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
