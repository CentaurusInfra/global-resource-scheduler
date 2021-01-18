load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["wait.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils/wait",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/utils/clock:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/runtime:go_default_library",
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
