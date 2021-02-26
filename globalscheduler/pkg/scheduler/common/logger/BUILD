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
        "//globalscheduler/pkg/scheduler/common/logger/trace:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
