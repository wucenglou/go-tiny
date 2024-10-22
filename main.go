package main

import (
	"context"
	"go-tiny/controller"
	"go-tiny/initialize"
	"go-tiny/routes"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	initialize.InitConfig()
	initialize.InitRedis()
	initialize.InitRabbitMQ()

	go controller.StartWorker()
	// initialize.InitElastic()
	// 初始化数据库连接
	initialize.DB = initialize.InitDB()
	if initialize.DB != nil {
		db, _ := initialize.DB.DB()
		defer db.Close()
	}

	// var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	// 导出博客条目
	// 	blogs, err := controller.ExportBlogs(initialize.DB)
	// 	if err != nil {
	// 		log.Printf("Error exporting blogs: %s", err)
	// 		return
	// 	}

	// 	// 使用 goroutines 并发处理每个博客条目，并限制最大并发数
	// 	concurrentLimit := 200 // 最大并发数量
	// 	semaphore := make(chan struct{}, concurrentLimit)
	// 	blogWg := sync.WaitGroup{}

	// 	blogWg.Add(len(blogs))

	// 	// 使用 goroutines 并发处理每个博客条目
	// 	for _, blog := range blogs {
	// 		go func(blog model.Blog) {
	// 			defer blogWg.Done()
	// 			semaphore <- struct{}{}        // 获取信号量
	// 			defer func() { <-semaphore }() // 释放信号量
	// 			if err := controller.IndexBlogToElasticsearch(blog); err != nil {
	// 				log.Printf("Error indexing blog to Elasticsearch: %s", err)
	// 			}
	// 		}(blog)
	// 	}

	// 	// 等待所有 goroutines 完成
	// 	blogWg.Wait()

	// 	log.Println("All blogs indexed successfully.")
	// }()

	// go func() {
	// 	wg.Wait()
	// 	log.Println("Data import completed.")
	// }()

	initialize.RedisClient.Set(context.Background(), "test", "Hello, Redis!", 0)
	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run(":" + viper.GetString("server.port"))
}
