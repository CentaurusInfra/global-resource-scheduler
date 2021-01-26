package(default_visibility = ["//visibility:public"])

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
        "//globalscheduler/cmd:all-srcs",
        "//globalscheduler/controllers/cluster:all-srcs",
        "//globalscheduler/controllers/dispatcher:all-srcs",
        "//globalscheduler/controllers/distributor:all-srcs",
        "//globalscheduler/controllers/scheduler:all-srcs",
        "//globalscheduler/controllers/util:all-srcs",
        "//globalscheduler/grpc/cluster:all-srcs",
        "//globalscheduler/pkg/apis/cluster:all-srcs",
        "//globalscheduler/pkg/apis/dispatcher:all-srcs",
        "//globalscheduler/pkg/apis/distributor:all-srcs",
        "//globalscheduler/pkg/apis/scheduler:all-srcs",
        "//globalscheduler/pkg/scheduler:all-srcs",
        "//globalscheduler/test/data_createdelete_tool:all-srcs",
    ],
    tags = ["automanaged"],
)
