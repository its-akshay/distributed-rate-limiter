package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/its-akshay/distributed-rate-limiter/internal/config"
	"github.com/its-akshay/distributed-rate-limiter/internal/database"
	"github.com/its-akshay/distributed-rate-limiter/internal/handler"
	"github.com/its-akshay/distributed-rate-limiter/internal/repository"
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
	ruleHandler := handler.NewRuleHandler(repo)
	router := gin.Default()

	router.POST("/rules", ruleHandler.CreateRule)
	router.GET("rules/:id", ruleHandler.GetRule)
	router.GET("rules", ruleHandler.ListRules)

	router.Run(":8080")

}


