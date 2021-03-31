load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "aggregate.go",
        "resource.go",
    ],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector/task",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/client/informers/internalinterfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
        "//globalscheduler/pkg/scheduler/common/constants:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//resourcecollector/pkg/collector/cloudclient:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/util/json:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/schedulerstats:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/aggregates:go_default_library",
        "//vendor/github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/hypervisors:go_default_library",
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
