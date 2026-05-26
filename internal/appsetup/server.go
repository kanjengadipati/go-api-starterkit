package appsetup

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"pleco-api/internal/ai"
	"pleco-api/internal/config"
	"pleco-api/internal/middleware"
	"pleco-api/internal/services"
	"pleco-api/internal/services/monitoring"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func RunAPI(registerDocs func(*gin.Engine)) error {
	config.LoadEnv()

	// Initialize structured logging
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	appConfig := config.LoadAppConfig()
	if err := appConfig.Validate(); err != nil {
		return err
	}

	db := config.ConnectDBWithDriver(appConfig.DatabaseURL, appConfig.DatabaseDriver)
	RunStartupTasks(appConfig, db)

	jwtService := services.NewJWTService(appConfig.JWTSecret)
	rateStore := newRateLimitStore(redisConnectionURL(appConfig))
	router, err := BuildRouter(db, appConfig, jwtService, rateStore)
	if err != nil {
		return err
	}

	// Initialize monitoring
	provider := os.Getenv("MONITORING_PROVIDER") // default: "none"
	baseMonitor, err := monitoring.NewMonitor(provider)
	if err != nil {
		slog.Warn("Monitoring initialization failed, falling back to no-op", "error", err)
		baseMonitor = &monitoring.NoOpMonitor{}
	}
	defer baseMonitor.Close()

	var monitor monitoring.Monitor = baseMonitor
	if os.Getenv("AI_MONITORING_ENABLED") == "true" {
		aiService, _ := ai.NewService(appConfig.AI)
		monitor = monitoring.NewAIMonitor(baseMonitor, aiService, db, true)
	}

	// Add monitoring middleware for error capture (5xx only)
	router.Use(func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 500 {
			err := fmt.Errorf(
				"HTTP %d on %s %s",
				c.Writer.Status(),
				c.Request.Method,
				c.Request.URL.Path,
			)
			monitor.CaptureException(err, c.Request.Context())
		}
	})

	if registerDocs != nil {
		registerDocs(router)
	}

	srv := &http.Server{
		Addr:              ":" + appConfig.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on :%s", appConfig.Port)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		return err
	case <-shutdownCtx.Done():
		log.Println("Shutdown signal received (SIGINT/SIGTERM)")
		log.Println("Gracefully shutting down HTTP server (max 10 seconds)...")
	}

	gracefulCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(gracefulCtx); err != nil {
		return err
	}

	if sqlDB, err := db.DB(); err == nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Warning: error closing database connection during shutdown: %v", err)
		}
	}

	if err := closeRateLimitStore(rateStore); err != nil {
		log.Printf("Error closing rate limit store: %v", err)
	}

	return <-serverErr
}

type rateLimitStoreCloser interface {
	middleware.RateLimitStore
	Close() error
}

func closeRateLimitStore(store middleware.RateLimitStore) error {
	closer, ok := store.(rateLimitStoreCloser)
	if !ok {
		return nil
	}
	return closer.Close()
}
