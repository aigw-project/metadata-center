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


package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDurationFromEnv(t *testing.T) {
	env := "TEST-ENV"
	os.Setenv(env, "")
	DurationFromEnv(env, nil)
	os.Setenv(env, "invalid")
	DurationFromEnv(env, nil)
	os.Setenv(env, "2s")
	call := false
	DurationFromEnv(env, func(d time.Duration) {
		require.Equal(t, 2*time.Second, d)
		call = true
	})
	require.True(t, call)
}

func TestIntFromEnv(t *testing.T) {
	env := "TEST-ENV"
	os.Setenv(env, "")
	IntFromEnv(env, nil)
	os.Setenv(env, "string")
	IntFromEnv(env, nil)
	os.Setenv(env, "10")
	call := false
	IntFromEnv(env, func(i int) {
		require.Equal(t, 10, i)
		call = true
	})
	require.True(t, call)
}

func TestBoolFromEnv(t *testing.T) {
	env := "TEST-ENV"
	os.Setenv(env, "")
	BoolFromEnv(env, nil)
	os.Setenv(env, "string")
	BoolFromEnv(env, nil)
	os.Setenv(env, "true")
	call := false
	BoolFromEnv(env, func(v bool) {
		require.True(t, v)
		call = true
	})
	require.True(t, call)
}
