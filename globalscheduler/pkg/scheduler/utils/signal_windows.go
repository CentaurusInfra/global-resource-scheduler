package utils

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
