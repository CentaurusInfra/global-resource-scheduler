load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "factory.go",
        "generic.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/apis/clusters/informers/externalversions",
    visibility = ["//visibility:public"],
    deps = [
        "//staging/src/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//staging/src/k8s.io/client-go/tools/cache:go_default_library",
        "//vendor/globalscheduler/pkg/apis/clusters/clientset/versioned:go_default_library",
        "//vendor/globalscheduler/pkg/apis/clusters/informers/externalversions/clusters:go_default_library",
        "//vendor/globalscheduler/pkg/apis/clusters/informers/externalversions/internalinterfaces:go_default_library",
        "//vendor/globalscheduler/pkg/apis/clusters/v1:go_default_library",
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
        "//globalscheduler/pkg/apis/clusters/informers/externalversions/clusters:all-srcs",
        "//globalscheduler/pkg/apis/clusters/informers/externalversions/internalinterfaces:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
