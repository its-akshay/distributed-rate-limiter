package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/config"
	"github.com/its-akshay/distributed-rate-limiter/internal/database"
	"github.com/its-akshay/distributed-rate-limiter/internal/handler"
	"github.com/its-akshay/distributed-rate-limiter/internal/limiter"
	"github.com/its-akshay/distributed-rate-limiter/internal/metrics"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
	"github.com/its-akshay/distributed-rate-limiter/internal/service"
)

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

	router.POST("/rules", ruleHandler.CreateRule)
	router.GET("/rules/:id", ruleHandler.GetRule)
	router.GET("/rules", ruleHandler.ListRules)
	router.POST("/check", ruleHandler.Check)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.Run(":8080")

}
