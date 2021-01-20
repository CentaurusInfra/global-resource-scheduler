load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["dispatcher_process.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/cmd/dispatcher_process",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/controllers/dispatcher:go_default_library",
        "//globalscheduler/controllers/util:go_default_library",
        "//staging/src/k8s.io/client-go/tools/clientcmd:go_default_library",
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

go_binary(
    name = "dispatcher_process",
    embed = [":go_default_library"],
    visibility = ["//visibility:public"],
)
