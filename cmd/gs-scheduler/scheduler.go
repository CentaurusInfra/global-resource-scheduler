package main

import (
	"fmt"
	"os"

	"k8s.io/kubernetes/cmd/gs-scheduler/app"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/options"
	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/utils"
)

func main() {
	s := options.NewServerRunOptions()
	stopCh := utils.SetupSignalHandler()
	if err := app.Run(s, stopCh); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
