load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["factory.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/informers",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/client:go_default_library",
        "//globalscheduler/pkg/scheduler/client/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/eipavailability:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/flavor:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/internalinterfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/siteinfos:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/siteresources:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/volumepool:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers/volumetype:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
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
        "//globalscheduler/pkg/scheduler/client/informers/eipavailability:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers/flavor:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers/internalinterfaces:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers/siteinfos:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers/siteresources:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers/volumepool:all-srcs",
        "//globalscheduler/pkg/scheduler/client/informers/volumetype:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
