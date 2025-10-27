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


package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/urfave/cli/v2"

	"github.com/aigw-project/metadata-center/pkg/config"
	"github.com/aigw-project/metadata-center/pkg/prom"
	"github.com/aigw-project/metadata-center/pkg/server"
	"github.com/aigw-project/metadata-center/pkg/utils/logger"
	"github.com/aigw-project/metadata-center/pkg/utils/version"
)

const AppName = "metadata-center"

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = version.MetadataCenterVersion
	app.Commands = []*cli.Command{
		newRunCmd(),
	}

	if err := app.Run(os.Args); err != nil {
		logger.Errorf("failed to start metadata center, args=%v, err=%v", os.Args, err)
		panic(err.Error())
	}
}

func newRunCmd() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Run server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "App configuration file(.json,.yaml,.toml)",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			return start(c.String("config"))
		},
	}
}

func start(configFile string) error {
	if err := config.Load(configFile); err != nil {
		logger.Errorf("failed to load config: %v", err)
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	config.InitEnv()

	srv := server.NewServer()
	srv.Init()

	prom.MetacenterNodeAlive.Set(1)
	if err := srv.Run(); err != nil {
		logger.Errorf("server runtime error: %v", err)
		return err
	}
	return nil
}
