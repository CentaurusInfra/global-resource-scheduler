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
        "//resourcecollector/pkg/collector/rpc/impl:all-srcs",
        "//resourcecollector/pkg/collector/rpc/pbfile:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
