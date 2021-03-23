load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["flavor.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers/flavor",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/client:go_default_library",
        "//globalscheduler/pkg/scheduler/client/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/internalinterfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/flavors:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
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
