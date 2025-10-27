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

package load

import (
	"encoding/json"
	"fmt"

	"github.com/aigw-project/metadata-center/pkg/replicator"
)

// Constants for replicator message types
const (
	// LoadStatsSet is the message type for setting load statistics
	LoadStatsSet = "load.stats.set"
	// LoadStatsDelete is the message type for deleting load statistics
	LoadStatsDelete = "load.stats.delete"
	// LoadPromptDelete is the message type for deleting prompt statistics
	LoadPromptDelete = "load.prompt.delete"
)

// init registers the load statistics handlers with the replicator
func init() {
	replicator.Register(LoadStatsSet, HandleLoadSet)
	replicator.Register(LoadStatsDelete, HandleLoadDelete)
	replicator.Register(LoadPromptDelete, HandleLoadPromptDelete)
}

// HandleLoadSet processes load statistics set messages
func HandleLoadSet(payload json.RawMessage) error {
	var req InferenceRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal payload for handleLoadSet: %w", err)
	}

	loadStats.AddRequest(&req)
	return nil
}

// HandleLoadDelete processes load statistics delete messages
func HandleLoadDelete(payload json.RawMessage) error {
	var req DeletionInferenceRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal payload for handleLoadDelete: %w", err)
	}

	loadStats.DeleteRequest(&req)
	return nil
}

// HandleLoadPromptDelete processes prompt statistics delete messages
func HandleLoadPromptDelete(payload json.RawMessage) error {
	var req DeletionInferenceRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal payload for handleLoadPromptDelete: %w", err)
	}

	loadStats.DeletePromptLength(&req)
	return nil
}
