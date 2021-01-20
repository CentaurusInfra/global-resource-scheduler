load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["core.go"],
    importmap = "k8s.io/kubernetes/vendor/github.com/go-chassis/go-archaius/core",
    importpath = "github.com/go-chassis/go-archaius/core",
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
        "//vendor/github.com/go-chassis/go-archaius/core/cast:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/core/config-manager:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/core/event-system:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
