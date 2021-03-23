load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["collector.go"],
    importpath = "k8s.io/kubernetes/resourcecollector/pkg/collector",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/apis/scheduler/v1:go_default_library",
        "//globalscheduler/pkg/scheduler/client/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//resourcecollector/pkg/collector/cloudclient:go_default_library",
        "//resourcecollector/pkg/collector/common/config:go_default_library",
        "//resourcecollector/pkg/collector/internal/cache:go_default_library",
        "//resourcecollector/pkg/collector/rpcclient:go_default_library",
        "//resourcecollector/pkg/collector/siteinfo:go_default_library",
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
        "//resourcecollector/pkg/collector/cloudclient:all-srcs",
        "//resourcecollector/pkg/collector/common/apiserver:all-srcs",
        "//resourcecollector/pkg/collector/common/config:all-srcs",
        "//resourcecollector/pkg/collector/httpserver:all-srcs",
        "//resourcecollector/pkg/collector/internal/cache:all-srcs",
        "//resourcecollector/pkg/collector/router:all-srcs",
        "//resourcecollector/pkg/collector/rpcclient:all-srcs",
        "//resourcecollector/pkg/collector/rpcserver:all-srcs",
        "//resourcecollector/pkg/collector/siteinfo:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
