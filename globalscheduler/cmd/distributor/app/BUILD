load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["command.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/cmd/distributor/app",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/cmd/distributor/app/config:go_default_library",
        "//globalscheduler/cmd/distributor/app/options:go_default_library",
        "//pkg/util/flag:go_default_library",
        "//pkg/version:go_default_library",
        "//pkg/version/verflag:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//staging/src/k8s.io/apiserver/pkg/util/term:go_default_library",
        "//staging/src/k8s.io/component-base/cli/flag:go_default_library",
        "//staging/src/k8s.io/component-base/cli/globalflag:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
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
        "//globalscheduler/cmd/distributor/app/config:all-srcs",
        "//globalscheduler/cmd/distributor/app/options:all-srcs",
        "//globalscheduler/cmd/distributor/app/testing:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
