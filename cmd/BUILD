package(default_visibility = ["//visibility:public"])

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
        "//cmd/clicheck:all-srcs",
        "//cmd/controller-manager/app:all-srcs",
        "//cmd/flat-network-controller:all-srcs",
        "//cmd/gendocs:all-srcs",
        "//cmd/genkubedocs:all-srcs",
        "//cmd/genman:all-srcs",
        "//cmd/genswaggertypedocs:all-srcs",
        "//cmd/genutils:all-srcs",
        "//cmd/genyaml:all-srcs",
        "//cmd/gs-scheduler:all-srcs",
        "//cmd/hyperkube:all-srcs",
        "//cmd/importverifier:all-srcs",
        "//cmd/kube-apiserver:all-srcs",
        "//cmd/kube-controller-manager:all-srcs",
        "//cmd/kube-proxy:all-srcs",
        "//cmd/kube-scheduler:all-srcs",
        "//cmd/kubeadm:all-srcs",
        "//cmd/kubectl:all-srcs",
        "//cmd/kubelet:all-srcs",
        "//cmd/kubemark:all-srcs",
        "//cmd/linkcheck:all-srcs",
        "//cmd/resource-collector:all-srcs",
    ],
    tags = ["automanaged"],
)
