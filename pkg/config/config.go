// Copyright The AIGW Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package config

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/koding/multiconfig"

	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// C is the global configuration instance
var (
	C = new(Config)
)

// InitEnv initializes environment-specific configurations
// Should be called after config.Load
func InitEnv() {
	if C.PProf.Enable {
		go func() {
			pprofAddr := fmt.Sprintf("%s:%d", C.PProf.Host, C.PProf.Port)
			logger.Infof("start pprof with %s", pprofAddr)
			logger.Errorf("pprof serve exit: %v", http.ListenAndServe(pprofAddr, nil))
		}()
	}
	// Parse environment variable configurations
	for _, es := range envSetters {
		es.setter(es.env)
	}
}

// Load loads configuration from file (toml/json/yaml)
// Supports environment variable overrides
func Load(file string) error {
	loaders := []multiconfig.Loader{
		&multiconfig.TagLoader{},
		&multiconfig.EnvironmentLoader{},
	}

	if strings.HasSuffix(file, "toml") {
		loaders = append(loaders, &multiconfig.TOMLLoader{Path: file})
	}
	if strings.HasSuffix(file, "json") {
		loaders = append(loaders, &multiconfig.JSONLoader{Path: file})
	}
	if strings.HasSuffix(file, "yaml") {
		loaders = append(loaders, &multiconfig.YAMLLoader{Path: file})
	}

	m := multiconfig.DefaultLoader{
		Loader:    multiconfig.MultiLoader(loaders...),
		Validator: multiconfig.MultiValidator(&multiconfig.RequiredValidator{}),
	}
	return m.Load(C)
}

// Config holds all application configuration settings
type Config struct {
	HTTP  HTTP  // HTTP server configuration
	PProf PProf // Profiling configuration
	Log   Log   // Logging configuration
}

// PProf configuration for performance profiling
type PProf struct {
	Enable bool   // Enable pprof server
	Host   string // PProf server host
	Port   int    // PProf server port
}

// HTTP configuration for server settings
type HTTP struct {
	Host     string // Server host address
	Port     int    // Server port
	CertFile string // TLS certificate file path
	KeyFile  string // TLS private key file path
}

// Log configuration for logging settings
type Log struct {
	Level         int    // Log level (0-5: TRACE, DEBUG, INFO, WARN, ERROR, FATAL)
	Format        string // Log format (text, json)
	Output        string // Log output destination (stdout, stderr, file)
	OutputFile    string // Log file path when output is "file"
	RotationCount int    // Number of log files to keep
	RotationTime  int    // Log rotation interval in hours
	GinOutput     string // Gin framework log output destination
	GinOutputFile string // Gin framework log file path
}
