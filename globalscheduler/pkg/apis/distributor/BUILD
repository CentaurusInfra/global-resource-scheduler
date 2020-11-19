load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["register.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/apis/distributor",
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
        "//globalscheduler/pkg/apis/distributor/client/clientset/versioned:all-srcs",
        "//globalscheduler/pkg/apis/distributor/client/informers/externalversions:all-srcs",
        "//globalscheduler/pkg/apis/distributor/client/listers/distributor/v1:all-srcs",
        "//globalscheduler/pkg/apis/distributor/client/listers/scheduler/v1:all-srcs",
        "//globalscheduler/pkg/apis/distributor/v1:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
