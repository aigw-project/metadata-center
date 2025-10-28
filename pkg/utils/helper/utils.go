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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

// MetaDataCenterPodIp is the environment variable name for pod IP address
// Used in multi-instance synchronization to exclude the instance's own IP
const MetaDataCenterPodIp = "POD_IP"

// IPAM manages IP address allocation within a CIDR block
type IPAM struct {
	base  uint32     // Base IP address as uint32
	index uint32     // Current allocation index
	ipNet *net.IPNet // IP network CIDR block
}

// NewIPAM creates a new IP address manager for the given CIDR block
// Example: "192.168.0.0/16"
func NewIPAM(base string) *IPAM {
	ip, ipnet, err := net.ParseCIDR(base)
	if err != nil {
		panic(err)
	}
	return &IPAM{
		base:  binary.BigEndian.Uint32(ip.To4()),
		index: 0,
		ipNet: ipnet,
	}
}

// Alloc allocates the next available IP address from the CIDR block
// Returns error if allocation exceeds the CIDR block range
func (a *IPAM) Alloc() (string, error) {
	idx := atomic.AddUint32(&a.index, 1)
	ip32 := a.base + idx
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, ip32)
	ip := net.IP(b)
	if !a.ipNet.Contains(ip) {
		return "", errors.New("ip alloc too much")
	}
	return ip.String(), nil
}

// JSONDuration is a time.Duration that supports JSON marshaling/unmarshaling
type JSONDuration time.Duration

// UnmarshalJSON parses duration from JSON string format
func (d *JSONDuration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = JSONDuration(dur)
	return nil
}

// MarshalJSON converts duration to JSON string format
func (d JSONDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// GetIntFromEnv retrieves integer value from environment variable
// Returns defaultValue if environment variable is not set or invalid
func GetIntFromEnv(name string, defaultValue int) int {
	env := os.Getenv(name)
	if env != "" {
		if d, err := strconv.Atoi(env); err == nil {
			return d
		}
	}
	return defaultValue
}

// GetDurationFromEnv retrieves duration value from environment variable
// Returns defaultValue if environment variable is not set or invalid
func GetDurationFromEnv(name string, defaultValue time.Duration) time.Duration {
	env := os.Getenv(name)
	if env != "" {
		if d, err := time.ParseDuration(env); err == nil {
			return d
		}
	}
	return defaultValue
}

// GetTraceIDFromCtx extracts trace ID from context
// Returns empty string if trace ID is not present
func GetTraceIDFromCtx(ctx context.Context) string {
	if v, ok := ctx.Value(trace.TraceKey).(string); ok {
		return v
	}
	return ""
}

// GetLocalHosts retrieves and validates the pod IP from environment
// Panics if POD_IP is not set or invalid
func GetLocalHosts() (string, error) {
	podIP := os.Getenv(MetaDataCenterPodIp)
	if podIP == "" {
		return "", fmt.Errorf("environment variable %s is not set", MetaDataCenterPodIp)
	}

	ip := net.ParseIP(podIP)
	if ip == nil || ip.To4() == nil {
		return "", fmt.Errorf("invalid IP format in POD_IP env var: got %q", podIP)
	}

	return podIP, nil
}
