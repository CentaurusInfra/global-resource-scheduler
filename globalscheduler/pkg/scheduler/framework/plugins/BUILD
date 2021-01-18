load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["registry.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/plugins",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/framework/interfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/defaultbinder:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/exclusivesite:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/flavor:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/locationandoperator:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/network:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/queuesort:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/regionandaz:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/siteavailability:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/siteresources:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/stackaffinity:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins/volume:go_default_library",
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
        "//globalscheduler/pkg/scheduler/framework/plugins/defaultbinder:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/exclusivesite:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/flavor:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/locationandoperator:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/network:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/queuesort:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/regionandaz:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/siteavailability:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/siteresources:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/stackaffinity:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins/volume:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
