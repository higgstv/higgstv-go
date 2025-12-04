package main

// @title           HiggsTV API Server
// @version         1.0
// @description     HiggsTV API Server 的 Golang 實作版本，採用分層架構設計，提供完整的 RESTful API 服務。支援 MongoDB 和 SQLite 兩種資料庫，可透過配置切換。認證方式：Session-based authentication（Cookie-based）
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.higgstv.com
// @contact.email  support@higgstv.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @securityDefinitions.cookie  ApiAuth
// @in                          cookie
// @name                        higgstv_session
// @description                 Session-based authentication

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/higgstv/higgstv-go/internal/api"
	"github.com/higgstv/higgstv-go/internal/api/handlers"
	"github.com/higgstv/higgstv-go/internal/api/middleware"
	"github.com/higgstv/higgstv-go/internal/config"
	"github.com/higgstv/higgstv-go/internal/database"
	"github.com/higgstv/higgstv-go/internal/migration"
	"github.com/higgstv/higgstv-go/pkg/logger"
	"github.com/higgstv/higgstv-go/pkg/metrics"
	"github.com/higgstv/higgstv-go/pkg/session"

	swagFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/higgstv/higgstv-go/docs/swagger" // Swagger 文件
)

func main() {
	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 驗證配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// 初始化 Logger
	if err := logger.Init(cfg.Server.Env); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	logger.Logger.Info("Starting HiggsTV API Server",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// 初始化 Session
	session.Init(cfg.Session.Secret)

	// 連接資料庫（支援 MongoDB 和 SQLite）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 解析資料庫類型
	dbType, err := database.ParseDatabaseType(cfg.Database.Type)
	if err != nil {
		log.Fatalf("Invalid database type: %v", err)
	}

	// 建立資料庫連線
	db, err := database.NewDatabase(ctx, database.DatabaseConfig{
		Type:     dbType,
		URI:      cfg.Database.URI,
		Database: cfg.Database.Database,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(context.Background()); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// 測試連線
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	logger.Logger.Info("Database connected",
		zap.String("type", string(dbType)),
		zap.String("uri", cfg.Database.URI),
	)

	// 確保資料庫索引已建立
	if err := database.EnsureIndexesWithTimeout(db); err != nil {
		logger.Logger.Warn("Failed to ensure database indexes", zap.Error(err))
		// 不中斷啟動，但記錄警告
	} else {
		logger.Logger.Info("Database indexes ensured")
	}

	// 執行資料庫遷移
	if err := migration.RunMigrationsWithTimeout(db); err != nil {
		logger.Logger.Warn("Failed to run migrations", zap.Error(err))
		// 不中斷啟動，但記錄警告
	} else {
		logger.Logger.Info("Database migrations completed")
	}

	// 設定 Gin
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 中介層
	router.Use(middleware.RequestID())            // Request ID
	router.Use(middleware.APIVersionMiddleware()) // API 版本
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())
	if logger.Logger != nil {
		router.Use(middleware.RequestLogging(logger.Logger))
	}
	router.Use(metrics.PrometheusMiddleware()) // Prometheus 指標
	router.Use(middleware.RateLimit())         // Rate limiting

	// 健康檢查端點（不需要 rate limiting）
	router.GET("/health", handlers.HealthCheck(db))
	router.GET("/ready", handlers.ReadinessCheck(db))

	// Prometheus 指標端點
	router.GET("/metrics", handlers.Metrics())

	// Swagger 文件（僅開發環境）
	if cfg.Server.Env == "development" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swagFiles.Handler))
		logger.Logger.Info("Swagger documentation available at /swagger/index.html")
	}

	// 設定路由
	api.SetupRoutes(router, db, cfg)

	// 啟動伺服器
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// 優雅關閉
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	fmt.Printf("Server is running on port %s\n", port)

	// 等待中斷信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}
