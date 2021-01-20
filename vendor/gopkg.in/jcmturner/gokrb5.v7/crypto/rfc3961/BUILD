load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "encryption.go",
        "keyDerivation.go",
        "nfold.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3961",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3961",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/common:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/etype:go_default_library",
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
