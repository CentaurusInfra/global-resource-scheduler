load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "archaius.go",
        "configurationfactory.go",
        "options.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/go-chassis/go-archaius",
    importpath = "github.com/go-chassis/go-archaius",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/go-chassis/go-archaius/core:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/core/cast:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/core/config-manager:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/core/event-system:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/sources/commandline-source:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/sources/configcenter:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/sources/enviromentvariable-source:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/sources/file-source:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/sources/memory-source:go_default_library",
        "//vendor/github.com/go-chassis/go-archaius/sources/utils:go_default_library",
        "//vendor/github.com/go-chassis/go-chassis-config:go_default_library",
        "//vendor/github.com/go-mesh/openlogging:go_default_library",
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
        "//vendor/github.com/go-chassis/go-archaius/core:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/sources/commandline-source:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/sources/configcenter:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/sources/enviromentvariable-source:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/sources/file-source:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/sources/memory-source:all-srcs",
        "//vendor/github.com/go-chassis/go-archaius/sources/utils:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
