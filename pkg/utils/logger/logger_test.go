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

package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	formatter := customFormatter{}
	s, _ := formatter.Format(&logrus.Entry{
		Level:   logrus.DebugLevel,
		Message: "meeage",
		Data: logrus.Fields{
			"key": "value",
		},
	})
	assert.Equal(t, "0001-01-01T00:00:00Z [debug] meeage key=value \n", string(s))
}
