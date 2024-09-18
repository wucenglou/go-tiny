package initialize

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
)

var EsClient *elasticsearch.Client

func InitElastic() {
	// 创建 Elasticsearch 客户端
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}
	// 测试连接
	resp, err := client.Info(
		client.Info.WithContext(context.Background()),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.IsError() {
		var e map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&e)
		fmt.Println(e)
	} else {
		log.Println("Connected to Elasticsearch.")
	}
	EsClient = client
}
