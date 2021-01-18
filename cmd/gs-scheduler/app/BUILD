load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "server.go",
        "version.go",
    ],
    importpath = "k8s.io/kubernetes/cmd/gs-scheduler/app",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler:go_default_library",
        "//globalscheduler/pkg/scheduler/common/apiserver:go_default_library",
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/options:go_default_library",
        "//globalscheduler/pkg/scheduler/router:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//pkg/util/flag:go_default_library",
        "//staging/src/k8s.io/apiserver/pkg/server:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
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
