load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cluster.pb.go",
        "cluster_grpc.pb.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/grpc/cluster/proto",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/status:go_default_library",
        "//vendor/google.golang.org/protobuf/reflect/protoreflect:go_default_library",
        "//vendor/google.golang.org/protobuf/runtime/protoimpl:go_default_library",
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
