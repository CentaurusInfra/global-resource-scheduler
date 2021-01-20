load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "bitreader.go",
        "bitwriter.go",
        "blockdec.go",
        "blockenc.go",
        "blocktype_string.go",
        "bytebuf.go",
        "bytereader.go",
        "decoder.go",
        "decoder_options.go",
        "dict.go",
        "enc_base.go",
        "enc_better.go",
        "enc_dfast.go",
        "enc_fast.go",
        "encoder.go",
        "encoder_options.go",
        "framedec.go",
        "frameenc.go",
        "fse_decoder.go",
        "fse_encoder.go",
        "fse_predefined.go",
        "hash.go",
        "history.go",
        "seqdec.go",
        "seqenc.go",
        "snappy.go",
        "zstd.go",
    ],
    importmap = "k8s.io/kubernetes/vendor/github.com/klauspost/compress/zstd",
    importpath = "github.com/klauspost/compress/zstd",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/klauspost/compress/huff0:go_default_library",
        "//vendor/github.com/klauspost/compress/snappy:go_default_library",
        "//vendor/github.com/klauspost/compress/zstd/internal/xxhash:go_default_library",
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
        "//vendor/github.com/klauspost/compress/zstd/internal/xxhash:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//visibility:public"],
)
