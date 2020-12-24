load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["common.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/controllers/util",
    visibility = ["//visibility:public"],
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
        "//globalscheduler/controllers/util/consistenthashing:all-srcs",
        "//globalscheduler/controllers/util/openstack:all-srcs",
        "//globalscheduler/controllers/util/union:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
