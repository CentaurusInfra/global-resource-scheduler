load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "collector_interfaces.go",
        "factory_interfaces.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/internalinterfaces",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/client:go_default_library",
        "//globalscheduler/pkg/scheduler/client/cache:go_default_library",
        "//resourcecollector/pkg/collector/cloudclient:go_default_library",
        "//resourcecollector/pkg/collector/siteinfo:go_default_library",
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
