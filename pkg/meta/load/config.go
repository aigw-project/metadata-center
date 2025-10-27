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

import "time"

var (
	DefaultGCInterval            = 60 * time.Second
	DefaultRequestExpireDuration = 660 * time.Second
)

var (
	gcInterval            = DefaultGCInterval
	requestExpireDuration = DefaultRequestExpireDuration
)

// SetGCInterval sets the garbage collection interval
func SetGCInterval(d time.Duration) {
	gcInterval = d
}

// SetRequestExpireDuration sets the request expiration duration
func SetRequestExpireDuration(d time.Duration) {
	requestExpireDuration = d
}
