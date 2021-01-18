load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "client_claims.go",
        "client_info.go",
        "credentials_info.go",
        "device_claims.go",
        "device_info.go",
        "kerb_validation_info.go",
        "pac_type.go",
        "s4u_delegation_info.go",
        "signature_data.go",
        "supplemental_cred.go",
        "upn_dns_info.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/pac",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/pac",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/chksumtype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/keyusage:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/types:go_default_library",
        "//vendor/gopkg.in/jcmturner/rpc.v1/mstypes:go_default_library",
        "//vendor/gopkg.in/jcmturner/rpc.v1/ndr:go_default_library",
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
