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


package log

import (
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/aigw-project/metadata-center/pkg/config"
	"github.com/aigw-project/metadata-center/pkg/ginx"
	"github.com/aigw-project/metadata-center/pkg/utils/errors"
	"github.com/aigw-project/metadata-center/pkg/utils/json"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// InitLogger initializes the logging system with configuration
// Returns a cleanup function and error if initialization fails
func InitLogger() (func(), error) {
	c := config.C.Log
	logger.SetLevel(logger.Level(c.Level))
	logger.SetFormatter(c.Format)

	var file *rotatelogs.RotateLogs
	if c.Output != "" {
		switch c.Output {
		case "stdout":
			logger.SetOutput(os.Stdout)
		case "stderr":
			logger.SetOutput(os.Stderr)
		case "file":
			if name := c.OutputFile; name != "" {
				_ = os.MkdirAll(filepath.Dir(name), 0777)

				f, err := rotatelogs.New(name+".%Y-%m-%d-%H",
					rotatelogs.WithLinkName(name),
					rotatelogs.WithRotationTime(time.Duration(c.RotationTime)*time.Hour),
					rotatelogs.WithRotationCount(uint(c.RotationCount)))
				if err != nil {
					return nil, err
				}

				logger.SetOutput(f)
				file = f
			}
		}
	}

	logger.Infof("config: %s", json.MarshalToString(config.C))

	return func() {
		if file != nil {
			file.Close()
		}
	}, nil
}

// levelMap maps string log levels to logger.Level constants
var levelMap = map[string]logger.Level{
	"FATAL": logger.FatalLevel,
	"ERROR": logger.ErrorLevel,
	"WARN":  logger.WarnLevel,
	"INFO":  logger.InfoLevel,
	"DEBUG": logger.DebugLevel,
	"TRACE": logger.TraceLevel,
}

// levelList contains all valid log level strings
var levelList = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

// LevelParam represents the request parameter for updating log level
type LevelParam struct {
	Level string `json:"LevelParam" form:"LevelParam"`
}

// UpdateLogLevel handles HTTP requests to update the log level dynamically
func UpdateLogLevel(c *gin.Context) {
	param := LevelParam{}

	if err := ginx.ParseJSON(c, &param); err != nil {
		ginx.ResError(c, err)
		return
	}

	if l, exist := levelMap[param.Level]; exist {
		config.C.Log.Level = int(l)
		logger.SetLevel(l)
		logger.Infof("update log level to %s", param.Level)
		ginx.ResSuccess(c, "log level updated successfully")
		return
	}

	ginx.ResError(c, errors.InvalidInput("log LevelParam can only be %v", levelList))
}
