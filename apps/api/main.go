package main

import (
	"log"
	"os"

	"fukuoka-ai-api/delivery/controller"
	"fukuoka-ai-api/infrastructure/database"
	"fukuoka-ai-api/infrastructure/mlservice"
	infraRepo "fukuoka-ai-api/infrastructure/repository"
	"fukuoka-ai-api/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/fukuoka_ai.db"
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	mlServiceURL := os.Getenv("ML_SERVICE_URL")
	if mlServiceURL == "" {
		mlServiceURL = "http://localhost:8000"
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

	// 依存関係の注入
	tripRepo := infraRepo.NewTripRepository(db)
	mlService := mlservice.NewMLService(mlServiceURL)
	tripUsecase := usecase.NewTripUsecase(tripRepo, mlService)
	tripController := controller.NewTripController(tripUsecase)

	v1 := router.Group("/v1")
	{
		v1.POST("/trips", tripController.CreateTrip)
		v1.POST("/trips/:trip_id/recompute", tripController.RecomputeTrip)
		v1.GET("/shares/:share_id", tripController.GetShare)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
