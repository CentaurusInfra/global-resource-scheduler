load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "carrier.go",
        "gen.go",
        "wire.pb.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/wire",
    importpath = "github.com/openzipkin-contrib/zipkin-go-opentracing/wire",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/flag:go_default_library",
        "//vendor/github.com/openzipkin-contrib/zipkin-go-opentracing/types:go_default_library",
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
