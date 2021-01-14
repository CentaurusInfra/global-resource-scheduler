package main

import (
	"fmt"
	"os"

	"k8s.io/kubernetes/cmd/resource-collector/app"
)

func main() {
	command := app.NewResourceCollectorCommand()

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
