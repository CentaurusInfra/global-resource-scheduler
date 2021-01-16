load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "Authenticator.go",
        "AuthorizationData.go",
        "Cryptosystem.go",
        "HostAddress.go",
        "KerberosFlags.go",
        "PAData.go",
        "PrincipalName.go",
        "TypedData.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/types",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/types",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/jcmturner/gofork/encoding/asn1:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/asn1tools:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/addrtype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/asnAppTag:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/nametype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/patype:go_default_library",
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
