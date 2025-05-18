package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"feature-flags/internal/handlers"
	"feature-flags/internal/repository/mongodb"
	"feature-flags/internal/services"
)

func main() {
	// MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:password123@localhost:27017"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Initialize repositories
	db := client.Database("finbox")
	featureRepo := mongodb.NewFeatureRepository(db)
	dependencyRepo := mongodb.NewFeatureDependencyRepository(db)

	// Initialize services
	featureService := services.NewFeatureService(featureRepo, dependencyRepo)

	// Initialize handlers
	featureHandler := handlers.NewFeatureHandler(featureService)

	// Initialize router
	r := gin.Default()

	// Feature routes
	features := r.Group("/api/features")
	{
		features.POST("", featureHandler.CreateFeature)
		features.GET("/:id", featureHandler.GetFeatureStatus)
		features.POST("/:id/enable", featureHandler.EnableFeature)
		features.POST("/:id/disable", featureHandler.DisableFeature)
		features.POST("/dependencies", featureHandler.AddDependency)
	}

	// Create a server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// kill (no param) default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
