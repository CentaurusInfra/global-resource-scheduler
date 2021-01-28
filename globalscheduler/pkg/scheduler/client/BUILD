load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["clientset.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/client",
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
        "//globalscheduler/pkg/scheduler/client/cache:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers:all-srcs",
        "//globalscheduler/pkg/scheduler/client/typed:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
