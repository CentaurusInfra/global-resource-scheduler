load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "clientset.go",
        "doc.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/apis/cluster/client/clientset/versioned",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned/typed/cluster/v1:go_default_library",
        "//staging/src/k8s.io/client-go/discovery:go_default_library",
        "//staging/src/k8s.io/client-go/rest:go_default_library",
        "//staging/src/k8s.io/client-go/util/flowcontrol:go_default_library",
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
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned/fake:all-srcs",
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned/scheme:all-srcs",
        "//globalscheduler/pkg/apis/cluster/client/clientset/versioned/typed/cluster/v1:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
