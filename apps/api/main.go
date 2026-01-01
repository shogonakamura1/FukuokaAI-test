package main

import (
	"log"
	"os"
	"path/filepath"

	"fukuoka-ai-api/controllers"
	"fukuoka-ai-api/infra/service"
	"fukuoka-ai-api/usecase"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// .envファイルを読み込む（複数のパスを試す）
	envPaths := []string{
		filepath.Join("..", "..", ".env"),  // apps/api から実行する場合
		".env",                               // プロジェクトルートから実行する場合
		filepath.Join("..", ".env"),         // apps から実行する場合
	}
	
	loaded := false
	for _, envPath := range envPaths {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("Loaded .env file from: %s", envPath)
			loaded = true
			break
		}
	}
	
	if !loaded {
		// .envファイルが見つからない場合でも続行（環境変数が直接設定されているか、Docker Compose経由で渡されている可能性がある）
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	router := gin.Default()

	// CORS設定
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-Id")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// リコメンド機能の依存関係
	geocodingService := service.NewGeocodingService()
	nearbySearchService := service.NewNearbySearchService()
	placeDetailsService := service.NewPlaceDetailsService()
	routeService := service.NewRouteService()
	recommendUsecase := usecase.NewRecommendUsecase(geocodingService, nearbySearchService, placeDetailsService)
	resultUsecase := usecase.NewResultUsecase(geocodingService, placeDetailsService, routeService)
	recommendController := controllers.NewRecommendController(recommendUsecase)
	addController := controllers.NewAddController()
	resultController := controllers.NewResultController(resultUsecase)
	geocodingController := controllers.NewGeocodingController(geocodingService)

	// リコメンド機能のエンドポイント
	router.POST("/recommend", recommendController.Recommend)
	// 場所追加機能のエンドポイント
	router.POST("/add/:place_id", addController.AddPlace)
	// ルート提案機能のエンドポイント
	router.POST("/result", resultController.Result)
	// ジオコーディング機能のエンドポイント（場所名からplace_idを取得）
	router.POST("/geocoding", geocodingController.GetPlaceID)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
