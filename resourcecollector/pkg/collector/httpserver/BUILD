load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "api_get_snapshot.go",
        "api_register_scheduler.go",
    ],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector/httpserver",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/apis/scheduler/v1:go_default_library",
        "//globalscheduler/pkg/scheduler/utils:go_default_library",
        "//resourcecollector/pkg/collector:go_default_library",
        "//vendor/github.com/emicklei/go-restful:go_default_library",
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
