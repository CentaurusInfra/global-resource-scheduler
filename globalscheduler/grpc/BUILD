load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["grpc-adaptor.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/grpc",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/grpc/cluster:go_default_library",
        "//globalscheduler/pkg/apis/cluster/v1:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
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
    srcs = [
        ":package-srcs",
        "//globalscheduler/grpc/cluster:all-srcs",
        "//globalscheduler/grpc/cluster_client:all-srcs",
        "//globalscheduler/grpc/cluster_server:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
