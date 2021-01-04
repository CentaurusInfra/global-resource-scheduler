load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["dispatcher.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/cmd/dispatcher",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/controllers/dispatcher:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client/informers/externalversions:go_default_library",
        "//globalscheduler/pkg/apis/dispatcher/client/clientset/versioned:go_default_library",
        "//globalscheduler/pkg/apis/dispatcher/client/informers/externalversions:go_default_library",
        "//staging/src/k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes:go_default_library",
        "//staging/src/k8s.io/client-go/tools/clientcmd:go_default_library",
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
        "//globalscheduler/cmd/dispatcher/dispatcherprocess:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
