load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "APRep.go",
        "APReq.go",
        "KDCRep.go",
        "KDCReq.go",
        "KRBCred.go",
        "KRBError.go",
        "KRBPriv.go",
        "KRBSafe.go",
        "Ticket.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/gopkg.in/jcmturner/gokrb5.v7/messages",
    importpath = "gopkg.in/jcmturner/gokrb5.v7/messages",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/jcmturner/gofork/encoding/asn1:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/asn1tools:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/config:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/credentials:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/crypto:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/adtype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/asnAppTag:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/errorcode:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/flags:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/keyusage:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/msgtype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/nametype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/iana/patype:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/keytab:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/krberror:go_default_library",
        "//vendor/gopkg.in/jcmturner/gokrb5.v7/pac:go_default_library",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
