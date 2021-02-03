load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cache.go",
        "interface.go",
        "site_tree.go",
        "snapshot.go",
    ],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector/internal/cache",
    visibility = ["//resourcecollector/pkg/collector:__subpackages__"],
    deps = [
        "//globalscheduler/pkg/scheduler/client/informers:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/listers:go_default_library",
        "//globalscheduler/pkg/scheduler/sitecacheinfo:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//globalscheduler/pkg/scheduler/utils:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/sets:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/wait:go_default_library",
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
        "//resourcecollector/pkg/collector/internal/cache/fake:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
