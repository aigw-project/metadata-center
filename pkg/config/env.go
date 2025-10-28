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
	"strconv"
	"time"

	"github.com/aigw-project/metadata-center/pkg/meta/load"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

const (
	LoadGCInterval    = "METADATA_CENTER_LOAD_GC_INTERVAL"
	LoadRequestExpire = "METADATA_CENTER_LOAD_REQ_EXPIRE"
)

type EnvSetter struct {
	env    string
	setter func(string)
}

var envSetters = []EnvSetter{
	{LoadGCInterval, func(env string) {
		DurationFromEnv(env, load.SetGCInterval)
	}},
	{LoadRequestExpire, func(env string) {
		DurationFromEnv(env, load.SetRequestExpireDuration)
	}},
}

// DurationFromEnv reads duration value from environment variable
func DurationFromEnv(env string, f func(d time.Duration)) {
	v := os.Getenv(env)
	if v == "" {
		logger.Infof("environment variable %s not set", env)
		return
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		logger.Errorf("environment variable %s value %s is not a valid duration string", env, v)
		return
	}
	f(d)
	logger.Infof("environment variable %s value set to %s", env, d.String())
}

// IntFromEnv reads integer value from environment variable
func IntFromEnv(env string, f func(i int)) {
	v := os.Getenv(env)
	if v == "" {
		logger.Infof("environment variable %s not set", env)
		return
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		logger.Errorf("environment variable %s value %s is not a valid integer string", env, v)
		return
	}
	f(i)
	logger.Infof("environment variable %s value set to %d", env, i)
}

// BoolFromEnv reads boolean value from environment variable
func BoolFromEnv(env string, f func(t bool)) {
	v := os.Getenv(env)
	if v == "" {
		logger.Infof("environment variable %s not set", env)
		return
	}
	t, err := strconv.ParseBool(v)
	if err != nil {
		logger.Errorf("environment variable %s value %s is not a valid boolean string", env, v)
		return
	}
	f(t)
	logger.Infof("environment variable %s value set to %t", env, t)
}
