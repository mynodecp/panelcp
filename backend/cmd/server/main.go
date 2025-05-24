package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mynodecp/mynodecp/backend/internal/api"
	"github.com/mynodecp/mynodecp/backend/internal/auth"
	"github.com/mynodecp/mynodecp/backend/internal/config"
	"github.com/mynodecp/mynodecp/backend/internal/database"
	"github.com/mynodecp/mynodecp/backend/internal/middleware"
	"github.com/mynodecp/mynodecp/backend/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New()
	defer log.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Initialize Redis
	redisClient, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Initialize auth service
	authService := auth.NewService(db, redisClient, cfg.Auth)

	// Initialize API services
	apiServices := api.NewServices(db, redisClient, authService, log)

	// Start gRPC server
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.UnaryServerInterceptor(log)),
		grpc.StreamInterceptor(middleware.StreamServerInterceptor(log)),
	)

	// Register gRPC services
	api.RegisterServices(grpcServer, apiServices)

	// Start gRPC server in goroutine
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRPCPort))
	if err != nil {
		log.Fatal("Failed to listen for gRPC", zap.Error(err))
	}

	go func() {
		log.Info("Starting gRPC server", zap.Int("port", cfg.Server.GRPCPort))
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Create gRPC-Gateway mux
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	// Register gRPC-Gateway handlers
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := api.RegisterGatewayHandlers(ctx, mux, fmt.Sprintf("localhost:%d", cfg.Server.GRPCPort), opts); err != nil {
		log.Fatal("Failed to register gateway handlers", zap.Error(err))
	}

	// Create Gin router for HTTP server
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit())
	router.Use(middleware.Security())
	router.Use(middleware.Logging(log))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.Server.Version,
		})
	})

	// Serve static files for frontend
	router.Static("/static", "./frontend/dist/assets")
	router.StaticFile("/", "./frontend/dist/index.html")

	// Mount gRPC-Gateway
	router.Any("/api/*path", gin.WrapH(mux))

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Info("Starting HTTP server", zap.Int("port", cfg.Server.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("HTTP server forced to shutdown", zap.Error(err))
	}

	// Shutdown gRPC server
	grpcServer.GracefulStop()

	// Close database connections
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
	}

	// Close Redis connection
	redisClient.Close()

	log.Info("Servers shutdown complete")
}
