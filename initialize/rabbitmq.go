package initialize

import (
	"log"

	"github.com/streadway/amqp"
)

func InitRabbitMQ() (*amqp.Connection, error) {
	// rabbitmqConfig := viper.Get("rabbitmq")

	// // 构建连接字符串
	// connString := fmt.Sprintf("amqp://%s:%s@%s:%d/",
	// 	rabbitmqConfig.(map[string]interface{})["username"],
	// 	rabbitmqConfig.(map[string]interface{})["password"],
	// 	rabbitmqConfig.(map[string]interface{})["host"],
	// 	rabbitmqConfig.(map[string]interface{})["port"])
	// log.Println(connString)

	// conn, err := amqp.Dial(connString)
	// defer conn.Close()
	// if err != nil {
	// 	fmt.Println(err)
	// 	log.Fatalf("Failed to connect to RabbitMQ: %s", err)

	// }
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer conn.Close()
	return conn, err

}
