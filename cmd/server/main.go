// Package main provides the main entry point for the UltraFit server application.
//
// @title UltraFit API
// @version 1.0.0
// @description UltraFit 多租户权限管理系统 API
// @termsOfService https://ultrafit.com/terms/
// @contact.name UltraFit Team
// @contact.email admin@ultrafit.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/varluffy/shield/docs" // swagger docs
	"github.com/varluffy/shield/internal/routes"
	"github.com/varluffy/shield/internal/wire"
	"github.com/varluffy/shield/pkg/response"
	"github.com/varluffy/shield/pkg/validator"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "configs/config.dev.yaml", "配置文件路径")
	flag.Parse()

	// 初始化应用
	app, cleanup, err := wire.InitializeApp(*configPath)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	// 初始化验证器
	if err := validator.InitGlobalValidator(app.Config.App.Language); err != nil {
		app.Logger.Warn("Failed to initialize validator, using default language",
			zap.Error(err),
			zap.String("language", app.Config.App.Language),
		)
		// 使用默认语言重试
		if err := validator.InitGlobalValidator(validator.DefaultLanguage); err != nil {
			log.Fatalf("Failed to initialize validator with default language: %v", err)
		}
	}

	// 设置验证错误翻译器
	response.SetValidationErrorTranslator(validator.TranslateValidationError)

	// 设置路由
	router := routes.SetupRoutes(
		app.Config,
		app.Logger,
		app.UserHandler,
		app.CaptchaHandler,
		app.PermissionHandler,
		app.RoleHandler,
		app.FieldPermissionHandler,
		app.BlacklistHandler,
		app.AuthMiddleware,
		app.BlacklistAuthMiddleware,
		app.BlacklistLogMiddleware,
	)

	// 记录启动信息
	app.Logger.Info("Starting UltraFit server",
		zap.String("version", app.Config.App.Version),
		zap.String("environment", app.Config.App.Environment),
		zap.String("address", fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port)),
	)

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.Config.Server.Host, app.Config.Server.Port),
		Handler:      router,
		ReadTimeout:  app.Config.Server.ReadTimeout,
		WriteTimeout: app.Config.Server.WriteTimeout,
		IdleTimeout:  app.Config.Server.IdleTimeout,
	}

	// 启动服务器
	go func() {
		app.Logger.Info("Server starting",
			zap.String("address", server.Addr),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("Failed to start server",
				zap.Error(err),
			)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	app.Logger.Info("Server shutting down...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		app.Logger.Fatal("Server forced to shutdown",
			zap.Error(err),
		)
	}

	// 关闭数据库连接
	if sqlDB, err := app.DB.DB(); err == nil {
		sqlDB.Close()
	}

	app.Logger.Info("Server stopped")
}
