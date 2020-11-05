package options

import (
	"os"
	"path"
	"path/filepath"

	"k8s.io/kubernetes/pkg/scheduler/common/logger"

	"github.com/spf13/pflag"
)

// ServerRunOptions contains required server running options
type ServerRunOptions struct {
	ConfFile    string
	LogConfFile string
}

// NewServerRunOptions constructs a new ServerRunOptions if existed.
func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{}

	//default
	configDir := path.Join(filepath.Dir(os.Args[0]), "..", "conf")
	configDirAbs, err := filepath.Abs(configDir)
	if err != nil {
		logger.Errorf("Initial config file is not correct. %v", err.Error())
		return s
	}

	s.ConfFile = path.Join(configDirAbs, "conf-odin.yaml")
	s.LogConfFile = path.Join(configDirAbs, "logging.yaml")

	return s
}

// AddFlags add flags to FlagSet
func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.ConfFile, "conffile", s.ConfFile, "config file")
	fs.StringVar(&s.LogConfFile, "log-conffile", s.LogConfFile, "log config file")
}
