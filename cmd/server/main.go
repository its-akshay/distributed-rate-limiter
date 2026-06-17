package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/config"
	"github.com/its-akshay/distributed-rate-limiter/internal/database"
	"github.com/its-akshay/distributed-rate-limiter/internal/handler"
	"github.com/its-akshay/distributed-rate-limiter/internal/limiter"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
	"github.com/its-akshay/distributed-rate-limiter/internal/service"
)

func main() {
	cfg := config.Load()

	pg, err := database.NewPostgres(cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pg.Close()

	err = pg.Ping(context.Background())
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

	repo := repository.NewRuleRepository(pg)
	luaLimiter := limiter.NewLuaSlidingWindowLimiter(rdb)
	rateLimiterService := service.NewRateLimiterService(
		repo,
		luaLimiter,
	)
	ruleHandler := handler.NewRuleHandler(repo, rateLimiterService)
	router := gin.Default()

	// sw := limiter.NewSlidingWindowLimiter(rdb)

	router.POST("/rules", ruleHandler.CreateRule)
	router.GET("/rules/:id", ruleHandler.GetRule)
	router.GET("/rules", ruleHandler.ListRules)
	router.POST("/check", ruleHandler.Check)

	router.Run(":8080")

}
