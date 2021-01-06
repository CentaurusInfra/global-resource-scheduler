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
        "//globalscheduler/cmd/dispatcher:all-srcs",
        "//globalscheduler/cmd/distributor:all-srcs",
        "//globalscheduler/cmd/grpc-server:all-srcs",
        "//globalscheduler/cmd/gs-controllers:all-srcs",
        "//globalscheduler/cmd/scheduler:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
