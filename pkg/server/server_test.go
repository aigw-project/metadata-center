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

package server

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name   string
		expect *Server
	}{
		{
			name:   "create new server",
			expect: &Server{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewServer()
			assert.NotNil(t, result)
			assert.NotNil(t, result.Engine)
		})
	}
}

func TestServer_RealAddr(t *testing.T) {
	tests := []struct {
		name   string
		server *Server
		expect string
	}{
		{
			name:   "server with listener",
			server: &Server{ln: &mockListener{addr: "127.0.0.1:8080"}},
			expect: "127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.server.RealAddr()
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestServer_Stop(t *testing.T) {
	tests := []struct {
		name   string
		server *Server
	}{
		{
			name:   "stop server",
			server: &Server{ln: &mockListener{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				tt.server.Stop()
			})
		})
	}
}

// TestServer_Init is skipped as Init() delegates to other modules
// which should have their own unit tests

// mockListener implements net.Listener for testing
type mockListener struct {
	addr string
}

func (m *mockListener) Accept() (net.Conn, error) {
	return nil, nil
}

func (m *mockListener) Close() error {
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return &mockAddr{addr: m.addr}
}

// mockAddr implements net.Addr for testing
type mockAddr struct {
	addr string
}

func (m *mockAddr) Network() string {
	return "tcp"
}

func (m *mockAddr) String() string {
	return m.addr
}
