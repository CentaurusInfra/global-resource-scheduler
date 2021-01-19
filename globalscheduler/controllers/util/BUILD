load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "common.go",
        "klog.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/controllers/util",
    visibility = ["//visibility:public"],
    deps = [
        "//staging/src/k8s.io/apimachinery/pkg/util/wait:go_default_library",
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
        "//globalscheduler/controllers/util/consistenthashing:all-srcs",
        "//globalscheduler/controllers/util/openstack:all-srcs",
        "//globalscheduler/controllers/util/union:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["common_test.go"],
    embed = [":go_default_library"],
)
