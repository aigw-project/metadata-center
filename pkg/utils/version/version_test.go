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
	"testing"
)

func TestInitMetadataCenterVersion_Success(t *testing.T) {
	VersionFile = "/tmp/test_version_file"
	err := os.WriteFile(VersionFile, []byte("v1.2.3"), 0644)
	if err != nil {
		t.Fatalf("failed to create test version file: %v", err)
	}
	defer os.Remove(VersionFile)

	initMetadataCenterVersion()

	if MetadataCenterVersion != "v1.2.3" {
		t.Errorf("expected version to be 'v1.2.3', got '%s'", MetadataCenterVersion)
	}
}

func TestInitMetadataCenterVersion_FileNotFound(t *testing.T) {
	VersionFile = "/tmp/nonexistent_version_file"

	initMetadataCenterVersion()

	if MetadataCenterVersion != DefaultVersion {
		t.Errorf("expected version to be default '%s', got '%s'", DefaultVersion, MetadataCenterVersion)
	}
}

func TestInitMetadataCenterVersion_EmptyFile(t *testing.T) {
	VersionFile = "/tmp/test_version_file"
	err := os.WriteFile(VersionFile, []byte("   "), 0644) // Empty content
	if err != nil {
		t.Fatalf("failed to create test version file: %v", err)
	}
	defer os.Remove(VersionFile)

	initMetadataCenterVersion()

	if MetadataCenterVersion != DefaultVersion {
		t.Errorf("expected version to be default '%s', got '%s'", DefaultVersion, MetadataCenterVersion)
	}
}
