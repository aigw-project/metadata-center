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
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aigw-project/metadata-center/pkg/utils/trace"
)

// Logger Logrus
type Logger = logrus.Logger

// Entry logrus.Entry alias
type Entry = logrus.Entry

// Hook logrus.Hook alias
type Hook = logrus.Hook

type Level = logrus.Level

// Define logger level
const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type customFormatter struct {
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	timestamp := entry.Time.Format(time.RFC3339)
	fmt.Fprintf(b, "%s ", timestamp)
	fmt.Fprintf(b, "[%s] ", entry.Level.String())
	fmt.Fprintf(b, "%s ", entry.Message)

	for key, value := range entry.Data {
		if key != "time" {
			fmt.Fprintf(b, "%s=%v ", key, value)
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

// SetLevel Set logger level
func SetLevel(level Level) {
	logrus.SetLevel(level)
}

// SetFormatter Set logger output format (json/text)
func SetFormatter(format string) {
	switch format {
	case "json":
		logrus.SetFormatter(new(logrus.JSONFormatter))
	default:
		logrus.SetFormatter(new(customFormatter))
	}
}

// Define key
const (
	TraceID = "trace_id"
)

func FromTraceIDContext(ctx context.Context) string {
	v := ctx.Value(trace.TraceKey)
	if v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// WithContext Use context create entry
func WithContext(ctx context.Context) *Entry {
	fields := logrus.Fields{}

	if v := FromTraceIDContext(ctx); v != "" {
		fields[TraceID] = v
	}

	return logrus.WithContext(ctx).WithFields(fields)
}

// Define logrus alias
var (
	Tracef          = logrus.Tracef
	Debugf          = logrus.Debugf
	Infof           = logrus.Infof
	Warnf           = logrus.Warnf
	Errorf          = logrus.Errorf
	Fatalf          = logrus.Fatalf
	Panicf          = logrus.Panicf
	Printf          = logrus.Printf
	SetOutput       = logrus.SetOutput
	SetReportCaller = logrus.SetReportCaller
	StandardLogger  = logrus.StandardLogger
	ParseLevel      = logrus.ParseLevel
)
