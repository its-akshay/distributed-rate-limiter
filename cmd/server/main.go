package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
// @host localhost:8080
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

	router.POST("/rules", ruleHandler.CreateRule)
	router.GET("/rules/:id", ruleHandler.GetRule)
	router.GET("/rules", ruleHandler.ListRules)
	router.POST("/check", ruleHandler.Check)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.Run(":8080")

}
