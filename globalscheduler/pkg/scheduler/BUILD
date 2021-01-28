load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "eventhandlers.go",
        "scheduler.go",
    ],
    importpath = "k8s.io/kubernetes/globalscheduler/pkg/scheduler",
    visibility = ["//visibility:public"],
    deps = [
        "//globalscheduler/pkg/scheduler/algorithmprovider:go_default_library",
        "//globalscheduler/pkg/scheduler/client/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/client/informers:go_default_library",
        "//globalscheduler/pkg/scheduler/client/typed:go_default_library",
        "//globalscheduler/pkg/scheduler/common/config:go_default_library",
        "//globalscheduler/pkg/scheduler/common/constants:go_default_library",
        "//globalscheduler/pkg/scheduler/common/logger:go_default_library",
        "//globalscheduler/pkg/scheduler/factory:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/interfaces:go_default_library",
        "//globalscheduler/pkg/scheduler/framework/plugins:go_default_library",
        "//globalscheduler/pkg/scheduler/internal/cache:go_default_library",
        "//globalscheduler/pkg/scheduler/internal/queue:go_default_library",
        "//globalscheduler/pkg/scheduler/sitecacheinfo:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
        "//globalscheduler/pkg/scheduler/utils:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/wait:go_default_library",
        "//globalscheduler/pkg/scheduler/utils/workqueue:go_default_library",
        "//pkg/util/pod:go_default_library",
        "//staging/src/k8s.io/api/core/v1:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/types:go_default_library",
        "//staging/src/k8s.io/apimachinery/pkg/util/runtime:go_default_library",
        "//staging/src/k8s.io/client-go/informers:go_default_library",
        "//staging/src/k8s.io/client-go/informers/core/v1:go_default_library",
        "//staging/src/k8s.io/client-go/kubernetes:go_default_library",
        "//staging/src/k8s.io/client-go/tools/cache:go_default_library",
        "//staging/src/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//vendor/github.com/satori/go.uuid:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["eventhandlers_test.go"],
    embed = [":go_default_library"],
    deps = [
        "//globalscheduler/pkg/scheduler/internal/cache/fake:go_default_library",
        "//globalscheduler/pkg/scheduler/types:go_default_library",
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
        "//globalscheduler/pkg/scheduler/algorithmprovider:all-srcs",
        "//globalscheduler/pkg/scheduler/client:all-srcs",
        "//globalscheduler/pkg/scheduler/common/apiserver:all-srcs",
        "//globalscheduler/pkg/scheduler/common/config:all-srcs",
        "//globalscheduler/pkg/scheduler/common/constants:all-srcs",
        "//globalscheduler/pkg/scheduler/common/logger:all-srcs",
        "//globalscheduler/pkg/scheduler/common/plugins:all-srcs",
        "//globalscheduler/pkg/scheduler/factory:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/interfaces:all-srcs",
        "//globalscheduler/pkg/scheduler/framework/plugins:all-srcs",
        "//globalscheduler/pkg/scheduler/internal/cache:all-srcs",
        "//globalscheduler/pkg/scheduler/internal/queue:all-srcs",
        "//globalscheduler/pkg/scheduler/labels:all-srcs",
        "//globalscheduler/pkg/scheduler/listers:all-srcs",
        "//globalscheduler/pkg/scheduler/options:all-srcs",
        "//globalscheduler/pkg/scheduler/router:all-srcs",
        "//globalscheduler/pkg/scheduler/service:all-srcs",
        "//globalscheduler/pkg/scheduler/sitecacheinfo:all-srcs",
        "//globalscheduler/pkg/scheduler/types:all-srcs",
        "//globalscheduler/pkg/scheduler/utils:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
