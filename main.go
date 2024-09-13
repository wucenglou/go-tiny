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
	initialize.InitRedis()
	initialize.InitElastic()
	// 初始化数据库连接
	initialize.DB = initialize.InitDB()
	if initialize.DB != nil {
		db, _ := initialize.DB.DB()
		defer db.Close()
	}

	initialize.RedisClient.Set(context.Background(), "test", "Hello, Redis!", 0)
	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run(":" + viper.GetString("server.port"))
}
