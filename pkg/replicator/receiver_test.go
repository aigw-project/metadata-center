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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandleReplicateEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cleanHandlers := func() {
		handlers = make(map[string]EventHandler)
	}

	var receivedPayload json.RawMessage

	testCases := []struct {
		name                 string
		eventTypeHeader      string
		requestBody          string
		setupFunc            func()
		expectedStatus       int
		expectedBodyContains string
		verifyFunc           func(t *testing.T)
	}{
		{
			name:            "should handle a valid event successfully",
			eventTypeHeader: "test.event",
			requestBody:     `{"key":"value"}`,
			setupFunc: func() {
				Register("test.event", func(payload json.RawMessage) error {
					return nil
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                 "should return 400 if Event-Type header is missing",
			requestBody:          `{}`,
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "missing Event-Type header",
		},
		{
			name:                 "should return 400 for an unregistered event type",
			eventTypeHeader:      "unknown.event",
			requestBody:          `{}`,
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "Unsupported event type",
		},
		{
			name:            "should return 400 when handler returns an error",
			eventTypeHeader: "error.event",
			requestBody:     `{"error":"test"}`,
			setupFunc: func() {
				Register("error.event", func(payload json.RawMessage) error {
					return io.ErrUnexpectedEOF
				})
			},
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "handler execute error",
		},
		{
			name:            "should handle an empty request body successfully",
			eventTypeHeader: "empty.test",
			requestBody:     "",
			setupFunc: func() {
				Register("empty.test", func(payload json.RawMessage) error {
					return nil
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:            "should handle a valid JSON request body and pass it to the handler",
			eventTypeHeader: "json.test",
			requestBody:     `{"test":"value","array":[1,2,3]}`,
			setupFunc: func() {
				Register("json.test", func(payload json.RawMessage) error {
					receivedPayload = payload
					return nil
				})
			},
			expectedStatus: http.StatusOK,
			verifyFunc: func(t *testing.T) {
				assert.JSONEq(t, `{"test":"value","array":[1,2,3]}`, string(receivedPayload))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanHandlers()
			receivedPayload = nil

			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/replicate", strings.NewReader(tc.requestBody))
			if tc.eventTypeHeader != "" {
				c.Request.Header.Set("Event-Type", tc.eventTypeHeader)
			}

			HandleReplicateEvent(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedBodyContains != "" {
				assert.Contains(t, w.Body.String(), tc.expectedBodyContains)
			}

			if tc.verifyFunc != nil {
				tc.verifyFunc(t)
			}
		})
	}

	t.Run("should return 400 on request body read error", func(t *testing.T) {
		cleanHandlers()
		Register("test.event", func(payload json.RawMessage) error { return nil })

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/replicate", &failingReader{})
		c.Request.Header.Set("Event-Type", "test.event")

		HandleReplicateEvent(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid body")
	})
}

func TestRegister(t *testing.T) {
	cleanHandlers := func() {
		handlers = make(map[string]EventHandler)
	}

	t.Run("should register a handler successfully", func(t *testing.T) {
		cleanHandlers()

		count := len(handlers)
		testHandler := func(payload json.RawMessage) error { return nil }

		Register("test.event", testHandler)

		assert.Equal(t, count+1, len(handlers))
		assert.NotNil(t, handlers["test.event"])
	})

	t.Run("should not register a handler for an empty event type", func(t *testing.T) {
		cleanHandlers()

		count := len(handlers)
		testHandler := func(payload json.RawMessage) error { return nil }

		Register("", testHandler)

		assert.Equal(t, count, len(handlers))
	})

	t.Run("should not register a nil handler", func(t *testing.T) {
		cleanHandlers()

		count := len(handlers)

		Register("test.event", nil)

		assert.Equal(t, count, len(handlers))
	})

	t.Run("should not overwrite an existing handler for the same event type", func(t *testing.T) {
		cleanHandlers()

		var handlerCalled bool
		firstHandler := func(payload json.RawMessage) error {
			handlerCalled = true
			return nil
		}
		secondHandler := func(payload json.RawMessage) error { return nil }

		Register("duplicate.event", firstHandler)
		count := len(handlers)

		Register("duplicate.event", secondHandler)

		assert.Equal(t, count, len(handlers))
		assert.NotNil(t, handlers["duplicate.event"])

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/replicate", bytes.NewBufferString("test"))
		c.Request.Header.Set("Event-Type", "duplicate.event")

		handlerCalled = false
		HandleReplicateEvent(c)
		assert.True(t, handlerCalled)
	})

	t.Run("should register multiple different handlers", func(t *testing.T) {
		cleanHandlers()

		handler1 := func(payload json.RawMessage) error { return nil }
		handler2 := func(payload json.RawMessage) error { return nil }
		handler3 := func(payload json.RawMessage) error { return nil }

		Register("event.1", handler1)
		Register("event.2", handler2)
		Register("event.3", handler3)

		assert.Equal(t, 3, len(handlers))
		assert.NotNil(t, handlers["event.1"])
		assert.NotNil(t, handlers["event.2"])
		assert.NotNil(t, handlers["event.3"])
	})
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}
