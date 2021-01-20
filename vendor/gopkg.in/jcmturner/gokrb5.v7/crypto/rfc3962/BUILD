load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "encryption.go",
        "keyDerivation.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3962",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3962",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/jcmturner/gofork/x/crypto/pbkdf2:go_default_library",
        "//vendor/gopkg.in/jcmturner/aescts.v1:go_default_library",
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
