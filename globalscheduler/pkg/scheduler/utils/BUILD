load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cmd.go",
        "constants.go",
        "error.go",
        "error_channel.go",
        "heap.go",
        "httputil.go",
        "rest_helper.go",
        "signal.go",
        "signal_posix.go",
        "signal_windows.go",
        "utils.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/common/constants:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//pkg/scheduler/metrics:go_default_library",
        "//staging/src/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/github.com/emicklei/go-restful:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
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
        "//globalscheduler/pkg/scheduler/utils/clock:all-srcs",
        "//globalscheduler/pkg/scheduler/utils/runtime:all-srcs",
        "//globalscheduler/pkg/scheduler/utils/sets:all-srcs",
        "//globalscheduler/pkg/scheduler/utils/wait:all-srcs",
        "//globalscheduler/pkg/scheduler/utils/workqueue:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
