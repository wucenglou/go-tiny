package main

import (
	"context"
	"fmt"
	"go-tiny/controller"
	"go-tiny/initialize"
	"go-tiny/routes"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/viper"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	initialize.InitConfig()
	initialize.InitRedis()

	// mq测试
	mq, err := initialize.InitRabbitMQ()
	if err != nil {
		fmt.Println(err)
	}
	ch, err := mq.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	q, err := ch.QueueDeclare(
		"blog_queue", // 队列名称
		false,        // 是否持久化
		false,        // 是否自动删除
		false,        // 是否独占队列
		false,        // 是否阻塞
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	body := "Hello World!"
	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		fmt.Println(err)
	}
	log.Printf(" [x] Sent %s\n", body)
	time.Sleep(time.Second * 1)
	msgs, err := ch.Consume(
		"blog_queue", // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	failOnError(err, "Failed to register a consumer")
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

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
