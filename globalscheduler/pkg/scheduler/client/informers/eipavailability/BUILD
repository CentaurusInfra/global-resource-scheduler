load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["eipavailability.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/eipavailability",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/client:go_default_library",
        "//globalscheduler/pkg/scheduler/client/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/internalinterfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
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
