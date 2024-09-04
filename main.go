package main

import (
	"context"
	"go-tiny/initialize"
	"go-tiny/routes"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	initialize.InitConfig()
	initialize.InitDB()

	initialize.InitRedis()
	initialize.RedisClient.Set(context.Background(), "test", "Hello, Redis!", 0)

	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run(":" + viper.GetString("server.port"))
}
