load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["distributor_server.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/cmd/distributor",
    visibility = ["//visibility:private"],
    deps = [
        "//globalscheduler/controllers/distributor:go_default_library",
        "//globalscheduler/pkg/apis/distributor/client/clientset/versioned:go_default_library",
        "//globalscheduler/pkg/apis/distributor/client/informers/externalversions:go_default_library",
        "//staging/src/k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes:go_default_library",
        "//staging/src/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//staging/src/k8s.io/component-base/logs:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
    ],
)

go_binary(
    name = "distributor",
    embed = [":go_default_library"],
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
        "//globalscheduler/cmd/distributor/process:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
