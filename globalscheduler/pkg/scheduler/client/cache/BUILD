load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cache.go",
        "lister.go",
        "shared_informer.go",
        "thread_safe_store.go",
        "ttlCache.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/client/cache",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/wait:go_default_library",
        "//vendor/github.com/deckarep/golang-set:go_default_library",
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
