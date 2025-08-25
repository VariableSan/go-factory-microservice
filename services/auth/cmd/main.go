package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/VariableSan/go-factory-microservice/pkg/common/config"
	"github.com/VariableSan/go-factory-microservice/pkg/common/database"
	"github.com/VariableSan/go-factory-microservice/pkg/common/logger"
	"github.com/VariableSan/go-factory-microservice/pkg/common/redis"
	"github.com/VariableSan/go-factory-microservice/services/auth/internal/server"
	"github.com/VariableSan/go-factory-microservice/services/auth/internal/service"
)

func main() {
	// Load common configuration
	commonCfg := config.LoadConfig()

	// Initialize centralized logger
	logger := logger.NewLogger(logger.Config{
		ServiceName: "auth-service",
		Level:       commonCfg.LogLevel,
		Format:      commonCfg.LogFormat,
	})

	// Load auth-specific configuration
	authCfg := config.LoadAuthConfig()

	// Initialize database connection
	db, err := database.NewFromURL(authCfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Connected to database successfully")

	// Initialize Redis client
	var redisClient *redis.Client
	if authCfg.RedisURL != "" {
		redisClient, err = redis.NewClientFromURL(authCfg.RedisURL)
		if err != nil {
			logger.Warn("Failed to connect to Redis", "error", err)
		} else {
			logger.Info("Connected to Redis successfully")
		}
	}

	// Initialize auth service
	authService := service.NewAuthService(db, authCfg.JWTSecret, redisClient, authCfg.TokenExpiry, authCfg.RefreshExpiry, logger)

	// Create servers
	httpServer := server.NewHTTPServer(authService, authCfg.HTTPPort, authCfg.JWTSecret, logger)
	grpcServer, err := server.NewGRPCServer(authService, authCfg.GRPCPort, logger)
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Start servers
	var wg sync.WaitGroup

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting HTTP server", "port", authCfg.HTTPPort)
		if err := httpServer.Start(); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Starting gRPC server", "port", authCfg.GRPCPort)
		if err := grpcServer.Start(); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop HTTP server
	go func() {
		if err := httpServer.Stop(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown failed", "error", err)
		}
	}()

	// Stop gRPC server
	go func() {
		grpcServer.Stop()
	}()

	// Close Redis connection
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error("Failed to close Redis connection", "error", err)
		}
	}

	// Close database connection
	if err := db.Close(); err != nil {
		logger.Error("Failed to close database connection", "error", err)
	} else {
		logger.Info("Database connection closed successfully")
	}

	wg.Wait()
	logger.Info("All servers stopped")
}
