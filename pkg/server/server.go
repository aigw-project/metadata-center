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


package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aigw-project/metadata-center/pkg/config"
	"github.com/aigw-project/metadata-center/pkg/log"
	"github.com/aigw-project/metadata-center/pkg/meta/load"
	"github.com/aigw-project/metadata-center/pkg/middleware"
	"github.com/aigw-project/metadata-center/pkg/replicator"
	"github.com/aigw-project/metadata-center/pkg/server/router"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
)

// Server represents the HTTP server with Gin engine and configuration
type Server struct {
	Engine *gin.Engine
	Config *config.Config
	ln     net.Listener
}

// NewServer creates and configures a new HTTP server
// Registers all API routes and middleware
func NewServer() *Server {
	engine := gin.New()
	engine.Use(middleware.GetMiddlewares()...)

	{
		g := engine.Group("")
		router.RegisterLogAPI(g)
		router.RegisterLoadAPI(g)
		router.RegisterStatusAPI(g)
		router.RegisterReplicateAPI(g)
	}

	return &Server{
		Engine: engine,
	}
}

// Run starts the HTTP server and listens for incoming requests
// Supports both HTTP and HTTPS based on configuration
func (s *Server) Run() error {
	cfg := config.C.HTTP
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.Engine,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
		IdleTimeout:  300 * time.Second,
	}

	// Use separate listener for easier unit testing
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		logger.Infof("server run as https config: %+v", cfg)
		return srv.ServeTLS(ln, cfg.CertFile, cfg.KeyFile)
	} else {
		logger.Infof("server run as http config: %+v", cfg)

		return srv.Serve(ln)
	}
}

// RealAddr returns the actual listening address of the server
func (s *Server) RealAddr() string {
	return s.ln.Addr().String()
}

// Stop gracefully shuts down the server by closing the listener
func (s *Server) Stop() {
	s.ln.Close()
}

// Init initializes server dependencies including logging and replication
func (s *Server) Init() {
	_, err := log.InitLogger()
	if err != nil {
		logger.Errorf("logger inint error: %v", err)
		return
	}

	load.Init()
	replicator.Init()
}
