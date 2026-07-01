package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/config"
	"github.com/its-akshay/distributed-rate-limiter/internal/database"
	"github.com/its-akshay/distributed-rate-limiter/internal/handler"
	"github.com/its-akshay/distributed-rate-limiter/internal/limiter"
	"github.com/its-akshay/distributed-rate-limiter/internal/metrics"
	"github.com/its-akshay/distributed-rate-limiter/internal/migration"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
	"github.com/its-akshay/distributed-rate-limiter/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/its-akshay/distributed-rate-limiter/docs"
)

// @title Distributed Rate Limiter API
// @version 1.0
// @description Distributed Rate Limiter using Go, Redis and PostgreSQL
// @host rate-limiter.local
// @BasePath /

func main() {

	cfg := config.Load()
	var pg *pgxpool.Pool
	var err error

	for i := 0; i < 10; i++ {
		pg, err = database.NewPostgres(cfg.PostgresURL)
		if err == nil {
			err = pg.Ping(context.Background())
			if err == nil {
				break
			}
		}

		log.Printf("Waiting for postgres... attempt %d", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PostgresSQL connected")
	wd, _ := os.Getwd()
	log.Println("Working directory:", wd)
	err = migration.Run(cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}

	rdb := database.NewRedis(cfg.RedisAddr)

	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Redis connected")

	metrics.Init()

	repo := repository.NewRuleRepository(pg)
	luaLimiter := limiter.NewLuaSlidingWindowLimiter(rdb)
	rateLimiterService := service.NewRateLimiterService(
		repo,
		luaLimiter,
	)
	ruleHandler := handler.NewRuleHandler(repo, rateLimiterService)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// sw := limiter.NewSlidingWindowLimiter(rdb)
	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	router.POST("/rules", ruleHandler.CreateRule)
	router.GET("/rules/:id", ruleHandler.GetRule)
	router.GET("/rules", ruleHandler.ListRules)
	router.POST("/check", ruleHandler.Check)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	healthHandler := handler.NewHealthHandler(pg, rdb)

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	go func() {
		log.Println("Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(
		quit,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	<-quit
	log.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	if err := rdb.Close(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}

	pg.Close()

	log.Println("Resources closed successfully")
}
