load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["client.go"],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector/rpcclient",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/grpc/cluster/proto:go_default_library",
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//resourcecollector/pkg/collector/common/config:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["client_test.go"],
    embed = [":go_default_library"],
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
