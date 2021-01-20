load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "aes128-cts-hmac-sha1-96.go",
        "aes128-cts-hmac-sha256-128.go",
        "aes256-cts-hmac-sha1-96.go",
        "aes256-cts-hmac-sha384-192.go",
        "crypto.go",
        "des3-cbc-sha1-kd.go",
        "rc4-hmac.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/crypto",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/crypto",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/crypto/md4:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/common:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/etype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3961:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3962:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc4757:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc8009:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/chksumtype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/etypeID:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/patype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/types:go_default_library",
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
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/common:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/etype:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3961:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc3962:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc4757:all-srcs",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto/rfc8009:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
