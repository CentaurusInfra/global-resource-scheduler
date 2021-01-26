load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["router.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/router",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/common/config:go_default_library",
        "//globalscheduler/pkg/scheduler/service:go_default_library",
        "//vendor/github.com/emicklei/go-restful:go_default_library",
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
