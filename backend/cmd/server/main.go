package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/heic-to-jpg/backend/internal/handler"
	"github.com/heic-to-jpg/backend/internal/infrastructure"
	"github.com/heic-to-jpg/backend/internal/usecase"
	vibelogger "github.com/ktanaha/vibe-coding-logger/pkg/logger"
)

func main() {
	// vibe-coding-loggerの初期化
	logger := vibelogger.Default()
	logger.Info("アプリケーション起動開始")

	// 依存関係の初期化
	converter := infrastructure.NewHeicConverter()
	uc := usecase.NewConvertImagesUsecase(converter)
	convertHandler := handler.NewConvertHandler(uc, logger)

	// 環境変数からポート番号を取得（デフォルト: 8080）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ginエンジンの初期化（デバッグモード）
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// CORS設定（開発環境用：すべてのオリジンを許可）
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length", "Content-Disposition"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// ロギングミドルウェア
	r.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		log.Printf("[REQUEST] %s %s from %s", method, path, c.ClientIP())

		tracker := logger.StartOperation(
			"HTTP リクエスト",
			map[string]interface{}{
				"method": method,
				"path":   path,
				"ip":     c.ClientIP(),
			},
		)

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[RESPONSE] %s %s - Status: %d - Duration: %s", method, path, status, duration)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("[ERROR] %s", err.Error())
			}
		}

		logger.CompleteOperation(tracker, map[string]interface{}{
			"status":   status,
			"duration": duration.String(),
		})
	})

	// ヘルスチェックエンドポイント
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 画像変換エンドポイント
	r.POST("/api/convert/single", convertHandler.ConvertSingle)
	r.POST("/api/convert/batch", convertHandler.ConvertBatch)
	r.POST("/api/convert/batch-zip", convertHandler.ConvertBatchZip)

	// サーバー起動
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
