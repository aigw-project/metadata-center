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

package helper

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

func TestNewIPAM(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		expectIP string
	}{
		{
			name:     "valid CIDR block",
			base:     "192.168.0.0/16",
			expectIP: "192.168.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := NewIPAM(tt.base)
			ip, err := ipam.Alloc()
			require.NoError(t, err)
			assert.Equal(t, tt.expectIP, ip)
		})
	}
}

func TestIPAM_Alloc(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		allocations int
		expectIPs   []string
	}{
		{
			name:        "single allocation",
			base:        "192.168.0.0/24",
			allocations: 1,
			expectIPs:   []string{"192.168.0.1"},
		},
		{
			name:        "multiple allocations",
			base:        "192.168.0.0/24",
			allocations: 3,
			expectIPs:   []string{"192.168.0.1", "192.168.0.2", "192.168.0.3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipam := NewIPAM(tt.base)
			for i := 0; i < tt.allocations; i++ {
				ip, err := ipam.Alloc()
				require.NoError(t, err)
				assert.Equal(t, tt.expectIPs[i], ip)
			}
		})
	}
}

func TestJSONDuration_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		expect JSONDuration
	}{
		{
			name:   "valid duration",
			data:   `"1h30m"`,
			expect: JSONDuration(90 * time.Minute),
		},
		{
			name:   "complex duration",
			data:   `"2h45m30s"`,
			expect: JSONDuration(2*time.Hour + 45*time.Minute + 30*time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d JSONDuration
			err := d.UnmarshalJSON([]byte(tt.data))
			require.NoError(t, err)
			assert.Equal(t, tt.expect, d)
		})
	}
}

func TestJSONDuration_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		duration JSONDuration
		expect   string
	}{
		{
			name:     "simple duration",
			duration: JSONDuration(30 * time.Minute),
			expect:   `"30m0s"`,
		},
		{
			name:     "hour duration",
			duration: JSONDuration(2 * time.Hour),
			expect:   `"2h0m0s"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.duration.MarshalJSON()
			require.NoError(t, err)
			assert.Equal(t, tt.expect, string(data))
		})
	}
}

func TestGetIntFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		value        string
		defaultValue int
		expect       int
	}{
		{
			name:         "valid integer",
			envVar:       "TEST_INT",
			value:        "42",
			defaultValue: 100,
			expect:       42,
		},
		{
			name:         "invalid integer",
			envVar:       "TEST_INVALID",
			value:        "not_a_number",
			defaultValue: 100,
			expect:       100,
		},
		{
			name:         "empty value",
			envVar:       "TEST_EMPTY",
			value:        "",
			defaultValue: 200,
			expect:       200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.value)
			result := GetIntFromEnv(tt.envVar, tt.defaultValue)
			assert.Equal(t, tt.expect, result)
			t.Setenv(tt.envVar, "")
		})
	}
}

func TestGetDurationFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		value        string
		defaultValue time.Duration
		expect       time.Duration
	}{
		{
			name:         "valid duration",
			envVar:       "TEST_DURATION",
			value:        "1h30m",
			defaultValue: 5 * time.Minute,
			expect:       90 * time.Minute,
		},
		{
			name:         "invalid duration",
			envVar:       "TEST_INVALID",
			value:        "invalid",
			defaultValue: 10 * time.Minute,
			expect:       10 * time.Minute,
		},
		{
			name:         "empty value",
			envVar:       "TEST_EMPTY",
			value:        "",
			defaultValue: 15 * time.Minute,
			expect:       15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.value)
			result := GetDurationFromEnv(tt.envVar, tt.defaultValue)
			assert.Equal(t, tt.expect, result)
			t.Setenv(tt.envVar, "")
		})
	}
}

func TestGetTraceIDFromCtx(t *testing.T) {
	tests := []struct {
		name   string
		ctx    context.Context
		expect string
	}{
		{
			name:   "context with trace ID",
			ctx:    context.WithValue(context.Background(), trace.TraceKey, "test-trace-id"),
			expect: "test-trace-id",
		},
		{
			name:   "context without trace ID",
			ctx:    context.Background(),
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTraceIDFromCtx(tt.ctx)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestGetLocalHosts(t *testing.T) {
	tests := []struct {
		name      string
		podIP     string
		expect    string
		expectErr bool
	}{
		{
			name:      "valid pod IP",
			podIP:     "192.168.1.1",
			expect:    "192.168.1.1",
			expectErr: false,
		},
		{
			name:      "invalid pod IP",
			podIP:     "invalid-ip",
			expect:    "",
			expectErr: true,
		},
		{
			name:      "empty pod IP",
			podIP:     "",
			expect:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(MetaDataCenterPodIp, tt.podIP)
			result, err := GetLocalHosts()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expect, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, result)
			}
			t.Setenv(MetaDataCenterPodIp, "")
		})
	}
}
