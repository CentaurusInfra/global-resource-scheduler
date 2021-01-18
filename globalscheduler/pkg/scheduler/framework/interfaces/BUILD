load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cycle_state.go",
        "framework.go",
        "interface.go",
        "registry.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/framework/interfaces",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/internal/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/listers:go_default_library",
        "//globalscheduler/pkg/scheduler/sitecacheinfo:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//globalscheduler/pkg/scheduler/utils:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/sets:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/workqueue:go_default_library",
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
