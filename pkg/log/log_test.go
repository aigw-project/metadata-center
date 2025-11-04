package log

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "path/filepath"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"

    "github.com/aigw-project/metadata-center/pkg/config"
)

func TestInitLogger(t *testing.T) {
    tests := []struct {
        name       string
        output     string
        useTempDir bool
        wantErr    bool
    }{
        {"stdout output", "stdout", false, false},
        {"stderr output", "stderr", false, false},
        {"file output no name", "file", false, false},  // no file path, should no write
        {"file output with name", "file", true, false}, // use temp dir
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // prepare config
            config.C.Log.Output = tt.output
            if tt.useTempDir {
                tmpDir := t.TempDir()
                config.C.Log.OutputFile = filepath.Join(tmpDir, "test.log")
            } else {
                config.C.Log.OutputFile = ""
            }
            config.C.Log.Level = int(levelMap["INFO"])
            config.C.Log.RotationTime = 1
            config.C.Log.RotationCount = 1

            cleanup, err := InitLogger()
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            if cleanup != nil {
                cleanup()
            }
        })
    }
}

func TestUpdateLogLevel(t *testing.T) {
    gin.SetMode(gin.TestMode)

    tests := []struct {
        name       string
        inputJSON  string
        wantStatus int
    }{
        {"valid level INFO", `{"LevelParam": "INFO"}`, http.StatusOK},
        {"invalid level", `{"LevelParam": "INVALID"}`, http.StatusBadRequest},
        {"missing level", `{}`, http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/loglevel", bytes.NewBufferString(tt.inputJSON))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()

            r := gin.New()
            r.POST("/loglevel", UpdateLogLevel)

            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)
        })
    }
}
