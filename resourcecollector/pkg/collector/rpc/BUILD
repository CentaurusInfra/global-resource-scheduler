load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["grpc_adaptor.go"],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector/rpc",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//resourcecollector/pkg/collector/common/config:go_default_library",
        "//resourcecollector/pkg/collector/rpc/impl:go_default_library",
        "//resourcecollector/pkg/collector/rpc/pbfile:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["grpc_adaptor_test.go"],
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
    srcs = [
        ":package-srcs",
        "//resourcecollector/pkg/collector/rpc/impl:all-srcs",
        "//resourcecollector/pkg/collector/rpc/pbfile:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
