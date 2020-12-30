load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["globalscheduler-controller-manager.go"],
    importpath = "k8s.io/kubernetes/cmd/globalscheduler",
    visibility = ["//visibility:private"],
    deps = [
        "//globalscheduler/controllers/cluster:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client/informers/externalversions:go_default_library",
        "//staging/src/k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes:go_default_library",
        "//staging/src/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
    ],
)

go_binary(
    name = "globalscheduler",
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
        "//cmd/globalscheduler/gs-dispatcher-process:all-srcs",
        "//cmd/globalscheduler/gs-distributor-controller:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
