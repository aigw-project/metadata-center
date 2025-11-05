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

package cache

import (
	"sync/atomic"
	"time"
)

// Default configuration constants
const (
	DefaultExpireDuration = time.Hour
	DefaultGcInterval     = 24 * time.Hour
	DefaultTopK           = 5
)

var (
	// Expiration duration for individual cache entries
	expireDuration = int64(DefaultExpireDuration)
	// GC interval for full pool traversal
	gcInterval = DefaultGcInterval
	// Default K value when TopK parameter is not provided
	defaultQueryTopK = DefaultTopK
	// Whether debug mode is enabled
	debugMode = atomic.Bool{}
)

// SetExpireDuration sets the expiration duration for cache entries
func SetExpireDuration(d time.Duration) {
	expireDuration = int64(d)
}

// SetGcInterval sets the garbage collection interval
func SetGcInterval(d time.Duration) {
	gcInterval = d
}

// SetDefaultQueryTopK sets the default TopK value for queries
func SetDefaultQueryTopK(k int) {
	defaultQueryTopK = k
}

// SetDebugMode enables or disables debug mode
func SetDebugMode(t bool) {
	debugMode.Store(t)
}

// IsDebugMode returns whether debug mode is enabled
func IsDebugMode() bool {
	return debugMode.Load()
}
