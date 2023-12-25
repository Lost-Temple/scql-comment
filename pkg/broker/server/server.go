// Copyright 2023 Ant Group Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/secretflow/scql/pkg/broker/application"
	"github.com/secretflow/scql/pkg/broker/constant"
	"github.com/secretflow/scql/pkg/broker/services/inter"
	"github.com/secretflow/scql/pkg/broker/services/intra"
)

type Server struct {
}

func NewIntraServer(app *application.App) (*http.Server, error) {
	router := gin.New()
	router.Use(ginLogger())
	router.Use(gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:80"},
		AllowMethods: []string{"GET", "PUT", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           50 * time.Second,
	}))
	// 内部接口，提供给本方的client调用
	intraSvc := intra.NewIntraSvc(app)
	router.POST(constant.DoQueryPath, intraSvc.DoQueryHandler)
	router.POST(constant.SubmitQueryPath, intraSvc.SubmitQueryHandler)
	router.POST(constant.FetchResultPath, intraSvc.FetchResultHandler)
	router.POST(constant.CreateProjectPath, intraSvc.CreateProjectHandler)
	router.POST(constant.ListProjectsPath, intraSvc.ListProjectsHandler)
	router.POST(constant.InviteMemberPath, intraSvc.InviteMemberHandler)
	router.POST(constant.ListInvitationsPath, intraSvc.ListInvitationsHandler)
	router.POST(constant.ProcessInvitationPath, intraSvc.ProcessInvitationHandler)
	router.POST(constant.CreateTablePath, intraSvc.CreateTableHandler)
	router.POST(constant.ListTablesPath, intraSvc.ListTablesHandler)
	router.POST(constant.DropTablePath, intraSvc.DropTableHandler)
	router.POST(constant.GrantCCLPath, intraSvc.GrantCCLHandler)
	router.POST(constant.RevokeCCLPath, intraSvc.RevokeCCLHandler)
	router.POST(constant.ShowCCLPath, intraSvc.ShowCCLHandler)
	router.POST(constant.EngineCallbackPath, intraSvc.EngineCallbackHandler)

	return &http.Server{
		Addr:           fmt.Sprintf("%s:%v", app.Conf.IntraServer.Host, app.Conf.IntraServer.Port),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}, nil
}

func NewInterServer(
	app *application.App) (*http.Server, error) {
	router := gin.New()
	router.Use(ginLogger())
	router.Use(gin.RecoveryWithWriter(logrus.StandardLogger().Out))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:80"},
		AllowMethods: []string{"GET", "PUT", "POST", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           50 * time.Second,
	}))
	// 外部接口，提供给对方的broker调用
	interSvc := inter.NewInterSvc(app)
	router.POST(constant.InviteToProjectPath, interSvc.InviteToProjectHandler)  // 邀请加入项目的服务接口
	router.POST(constant.ReplyInvitationPath, interSvc.ReplyInvitationHandler)  // 接收邀请结果的服务接口
	router.POST(constant.SyncInfoPath, interSvc.SyncInfoHandler)                // 同步信息服务接口，接收对方的同步信息请求：AddProjectMember、Create/Drop Table、Grant/Revoke CCL
	router.POST(constant.AskInfoPath, interSvc.AskInfoHandler)                  // 查询信息：提供项目、表、CCL的查询服务
	router.POST(constant.DistributeQueryPath, interSvc.DistributedQueryHandler) // 联合SQL查询接口
	router.POST(constant.ExchangeJobInfoPath, interSvc.ExchangeJobInfoHandler)  // 联合SQL查询时会调用对方的这个接口

	return &http.Server{
		Addr:           fmt.Sprintf("%s:%v", app.Conf.InterServer.Host, app.Conf.InterServer.Port),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}, nil
}

// ginLogger creates a gin logger middleware
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()

		msg := fmt.Sprintf("|GIN|status=%v|method=%v|path=%v|ip=%v|latency=%v|%s", status, c.Request.Method, path, c.ClientIP(), latency, c.Errors.String())
		if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
			logrus.Warn(msg)
		} else if status >= http.StatusInternalServerError {
			logrus.Error(msg)
		} else {
			logrus.Info(msg)
		}
	}
}
