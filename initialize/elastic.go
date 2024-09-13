package initialize

import (
	"context"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
)

var EsClient *elastic.Client

func InitElastic() {
	// 创建 Elasticsearch 客户端
	var err error
	esClient, err := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetSniff(false),
		elastic.SetHealthcheckInterval(time.Second*30),
	)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	// 检查 Elasticsearch 是否可用
	info, code, err := esClient.Ping("http://localhost:9200").Do(context.Background())
	if err != nil {
		log.Fatalf("Error pinging Elasticsearch: %v", err)
	}
	EsClient = esClient
	log.Printf("Elasticsearch returned with code %d and version %s", code, info.Version.Number)
}
