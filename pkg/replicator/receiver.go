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


package replicator

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/ginx"
	"github.com/aigw-project/metadata-center/pkg/utils/errors"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// EventHandler defines the function signature for handling replication events
type EventHandler func(payload json.RawMessage) error

// handlers stores registered event handlers by event type
var handlers = make(map[string]EventHandler)

// HandleReplicateEvent processes incoming replication events
// Validates event type, finds appropriate handler, and executes it
func HandleReplicateEvent(c *gin.Context) {
	eventType := c.GetHeader(EventTypeHeader)
	if eventType == "" {
		logger.Errorf("ReplicateAPI: missing Event-Type header")
		ginx.ResError(c, errors.InvalidInput("missing Event-Type header"))
		return
	}

	c.Set(EventTypeCtxKey, eventType)

	handler, ok := handlers[eventType]
	if !ok {
		logger.Errorf("ReplicateAPI: No handler found for event type: %s", eventType)
		ginx.ResError(c, errors.InvalidInput("Unsupported event type"))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Errorf("ReplicateAPI: read body error: %v", err)
		ginx.ResError(c, errors.InvalidInput("invalid body"))
		return
	}

	if err := handler(body); err != nil {
		logger.Errorf("ReplicateAPI: handler error: %v", err)
		ginx.ResError(c, errors.InvalidInput("handler execute error"))
		return
	}

	ginx.ResSuccess(c, nil)
}

// Register adds a new event handler for the specified event type
// Validates input parameters and prevents duplicate registrations
func Register(eventType string, handler EventHandler) {
	if eventType == "" {
		logger.Errorf("ReplicateAPI: missing Event-Type header")
		return
	}

	if handler == nil {
		logger.Errorf("ReplicateAPI: handler is nil")
		return
	}

	if _, exists := handlers[eventType]; exists {
		logger.Errorf("ReplicateAPI: handler already registered")
		return
	}

	handlers[eventType] = handler
	logger.Infof("ReplicateAPI: event handler registered: %s", eventType)
}
