load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["union.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/controllers/util/union",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/apis/cluster/v1:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/v1:go_default_library",
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

go_test(
    name = "go_default_test",
    srcs = ["union_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//globalscheduler/pkg/apis/cluster/v1:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/v1:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
    ],
)
