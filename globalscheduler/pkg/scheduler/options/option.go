/*
Copyright 2020 Authors of Arktos.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"os"
	"path"
	"path/filepath"

	"k8s.io/kubernetes/globalscheduler/pkg/scheduler/common/logger"

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
