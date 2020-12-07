load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["scheduler_process.go"],
    importpath = "k8s.io/kubernetes/globalscheduler/controllers/scheduler/process",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/controllers/util/union:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client:go_default_library",
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/client:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/client/clientset/versioned:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/client/clientset/versioned/scheme:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/client/informers/externalversions:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/client/informers/externalversions/scheduler/v1:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/client/listers/scheduler/v1:go_default_library",
        "//globalscheduler/pkg/apis/scheduler/v1:go_default_library",
        "//staging/src/k8s.io/api/core/v1:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/util/runtime:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/util/wait:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes/scheme:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes/typed/core/v1:go_default_library",
        "//staging/src/k8s.io/client-go/tools/cache:go_default_library",
        "//staging/src/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//staging/src/k8s.io/client-go/tools/record:go_default_library",
        "//staging/src/k8s.io/client-go/util/workqueue:go_default_library",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
