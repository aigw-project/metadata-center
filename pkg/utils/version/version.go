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

package version

import (
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// Version-related variables for metadata center
var (
	// MetadataCenterVersion holds the current version of the application
	MetadataCenterVersion string
	// VersionFile is the path to the version file
	VersionFile = "VERSION"
	// DefaultVersion is the fallback version if version file cannot be read
	DefaultVersion = "v0.0.0"
)

// init initializes the metadata center version on package load
func init() {
	initMetadataCenterVersion()
}

// initMetadataCenterVersion reads version from file and initializes metrics
// Falls back to default version if file cannot be read or is empty
func initMetadataCenterVersion() {
	data, err := os.ReadFile(VersionFile)
	if err != nil {
		logger.Errorf("failed to read version file %s: %v", VersionFile, err)
		MetadataCenterVersion = DefaultVersion
	} else {
		version := strings.TrimSpace(string(data))
		// Use default value if file is empty or contains only whitespace
		if version == "" {
			logger.Errorf("version file %s is empty or contains only whitespace", VersionFile)
			MetadataCenterVersion = DefaultVersion
		} else {
			MetadataCenterVersion = version
		}
	}

	// Set Prometheus metrics with the current version
	prom.AppVersionInfo.With(prometheus.Labels{"version": MetadataCenterVersion}).Set(1)
	logger.Infof("MetadataCenterVersion initialized to: %s", MetadataCenterVersion)
}
