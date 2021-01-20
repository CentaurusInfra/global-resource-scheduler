load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["scheduler.go"],
    importpath = "k8s.io/kubernetes/cmd/gs-scheduler",
    visibility = ["//visibility:private"],
    deps = [
        "//cmd/gs-scheduler/app:go_default_library",
        "//globalscheduler/pkg/scheduler/common/config:go_default_library",
    ],
)

go_binary(
    name = "gs-scheduler",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
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
        "//cmd/gs-scheduler/app:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
