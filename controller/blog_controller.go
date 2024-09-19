package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go-tiny/initialize"
	"go-tiny/model"
	"go-tiny/utils"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/util/rand"
)

type BlogController struct{}

var titles = []string{
	"如何在家制作美味蛋糕",
	"探索宇宙的奥秘",
	"编程入门指南",
	"健身与健康",
	"旅行的意义",
	// 更多标题...
}

func randomTitle() string {
	return titles[rand.Intn(len(titles))]
}

func randomContent() string {
	words := []string{"这是一个", "测试", "文章", "内容", "包括", "各种", "各样的", "句子", "为了", "填充", "博客"}
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < 100; i++ {
		b.WriteString(words[rand.Intn(len(words))])
		b.WriteString(" ")
	}
	return b.String()
}

//	for i := 0; i < 1000; i++ {
//		blog := model.Blog{
//			Title:       randomTitle(),
//			Content:     randomContent(),
//			AuthorID:    user.ID,
//			PublishedAt: time.Now(),
//		}
//		initialize.DB.Create(&blog)
//	}
func (bc *BlogController) CreateBlog(c *gin.Context) {
	var blog model.Blog
	if err := c.ShouldBindJSON(&blog); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 获取当前登录用户的 ID
	claims, _ := utils.GetClaims(c)
	currentUserID := claims.Id
	var user model.User
	if err := initialize.DB.Where("id = ?", currentUserID).First(&user).Error; err != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}
	blog.AuthorID = user.ID
	initialize.DB.Create(&blog)

	// go func() {
	// // 将博客条目索引到 Elasticsearch
	// esBlog := map[string]interface{}{
	// 	"id":        blog.ID,
	// 	"title":     blog.Title,
	// 	"content":   blog.Content,
	// 	"author_id": blog.AuthorID,
	// 	"author": map[string]interface{}{
	// 		"id":   user.ID,
	// 		"name": user.Username, // 假设 User 有一个 Username 字段
	// 	},
	// 	"published_at": blog.PublishedAt,
	// 	"created_at":   blog.CreatedAt,
	// 	"updated_at":   blog.UpdatedAt,
	// }

	// // 将文档序列化为 JSON 字符串
	// body, err := json.Marshal(esBlog)
	// if err != nil {
	// 	log.Printf("Failed to marshal blog document: %v", err)
	// }

	// // 索引文档到 Elasticsearch
	// req := esapi.IndexRequest{
	// 	Index:      "blogs",
	// 	DocumentID: fmt.Sprintf("%d", blog.ID), // 使用正确的字段名称 DocumentID
	// 	Body:       strings.NewReader(string(body)),
	// 	OpType:     "index",
	// 	Refresh:    "wait_for",
	// }
	// res, err := req.Do(context.Background(), initialize.EsClient)
	// if err != nil {
	// 	c.JSON(500, gin.H{"error": "Failed to index blog to Elasticsearch"})
	// 	return
	// }
	// defer res.Body.Close()
	// if res.IsError() {
	// 	log.Printf("[%s] Error indexing document: %s", res.Status(), res.String())
	// }
	// }()

	// 将索引任务推送到Redis队列
	enqueueIndexTask(blog, user)

	c.JSON(200, blog)
}

// 将索引任务推送到Redis队列
func enqueueIndexTask(blog model.Blog, user model.User) {
	esBlog := map[string]interface{}{
		"id":        blog.ID,
		"title":     blog.Title,
		"content":   blog.Content,
		"author_id": blog.AuthorID,
		"author": map[string]interface{}{
			"id":   user.ID,
			"name": user.Username, // 假设 User 有一个 Username 字段
		},
		"published_at": blog.PublishedAt,
		"created_at":   blog.CreatedAt,
		"updated_at":   blog.UpdatedAt,
	}

	body, err := json.Marshal(esBlog)
	if err != nil {
		log.Printf("Failed to marshal blog document: %v", err)
		return
	}

	ctx := context.Background()
	_, err = initialize.RedisClient.RPush(ctx, "blog_index_queue", body).Result()
	if err != nil {
		log.Printf("添加队列失败: %v", err)
		return
	}
	log.Printf("添加队列 ID %d", blog.ID)
}

func StartWorker() {
	ctx := context.Background()
	for {
		// 从Redis队列中取出任务
		result, err := initialize.RedisClient.BLPop(ctx, 0, "blog_index_queue").Result()
		if err != nil {
			log.Printf("Failed to dequeue task from Redis: %v", err)
			time.Sleep(time.Second * 5)
			continue
		}
		// 解析任务
		var esBlog map[string]interface{}
		err = json.Unmarshal([]byte(result[1]), &esBlog)
		// log.Printf("解析任务: %v", esBlog)
		if err != nil {
			log.Printf("Failed to parse task: %v", err)
			continue
		}
		EndexBlogToElasticsearch(esBlog)
	}
}

func (bc *BlogController) GetBlogs(c *gin.Context) {
	var blogs []model.Blog
	result := initialize.DB.Preload("Author").Find(&blogs)
	if result.Error != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch blogs"})
		return
	}
	c.JSON(200, blogs)
}

func (bc *BlogController) GetBlog(c *gin.Context) {
	var blog model.Blog
	if err := initialize.DB.Preload("Author").Where("id = ?", c.Param("id")).First(&blog).Error; err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Blog not found"})
		return
	}
	c.JSON(200, blog)
}

func (bc *BlogController) UpdateBlog(c *gin.Context) {
	var blog model.Blog
	if err := initialize.DB.Preload("Author").Where("id = ?", c.Param("id")).First(&blog).Error; err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "Blog not found"})
		return
	}

	// 确保只有该用户才能更新自己的博客
	if c.GetString("username") != blog.Author.Username {
		c.AbortWithStatusJSON(403, gin.H{"error": "Permission denied"})
		return
	}

	if err := c.ShouldBindJSON(&blog); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	initialize.DB.Save(&blog)

	go func() {
		// 构建更新博客的索引请求
		esBlog := map[string]interface{}{
			"id":        blog.ID,
			"title":     blog.Title,
			"content":   blog.Content,
			"author_id": blog.AuthorID,
			"author": map[string]interface{}{
				"id":   blog.Author.ID,
				"name": blog.Author.Username,
			},
			"published_at": blog.PublishedAt,
		}

		// 将文档序列化为 JSON 字符串
		body, err := json.Marshal(esBlog)
		if err != nil {
			log.Printf("Failed to marshal blog document: %v", err)
		}
		req := esapi.UpdateRequest{
			Index:      "blogs",
			DocumentID: fmt.Sprintf("%d", blog.ID),
			Body:       strings.NewReader(string(body)),
			Refresh:    "wait_for",
		}
		res, err := req.Do(context.Background(), initialize.EsClient)
		if err != nil {
			log.Printf("Failed to update blog in Elasticsearch: %v", err)
		}
		defer res.Body.Close()
	}()

	c.JSON(200, blog)
}

func (bc *BlogController) DeleteBlog(c *gin.Context) {
	var blog model.Blog
	fmt.Println("Deleting blog with ID:", c.Param("id"))
	id := c.Param("id")
	if initialize.DB.Where("id = ?", id).Delete(&blog).RowsAffected == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "Blog not found"})
		return
	}
	// 构建删除博客的索引请求
	req := esapi.DeleteRequest{
		Index:      "blogs",
		DocumentID: id,
	}
	res, err := req.Do(context.Background(), initialize.EsClient)
	if err != nil {
		log.Printf("Failed to delete blog from Elasticsearch: %v", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Printf("[%s] Error deleting document: %s", res.Status(), res.String())
	}

	c.JSON(200, gin.H{"message": "Blog deleted"})
}

func (bc *BlogController) SearchBlogs(c *gin.Context) {
	var blogs []model.Blog
	query := c.Query("keyword")
	initialize.DB.Where("title LIKE ? OR content LIKE ?", "%"+query+"%", "%"+query+"%").Find(&blogs)
	c.JSON(200, blogs)
}

func (bc *BlogController) SearchBlogsByElastic(c *gin.Context) {
	// 获取查询关键词
	query := c.Query("keyword")
	fmt.Println("Searching for:", query)

	q := esapi.SearchRequest{
		Index: []string{"blogs"},
		Body: strings.NewReader(`{
			"query": {
				"multi_match": {
					"query": "` + query + `",
					"fields": ["title^2", "content"],
					"analyzer": "ik_max_word",
					"type": "most_fields"
				}
			}
		}`),
	}
	res, err := q.Do(context.Background(), initialize.EsClient)
	if err != nil {
		fmt.Println("Error searching:", err)
	}
	defer res.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(res.Body).Decode(&result)
	// 返回查询结果
	c.JSON(200, result["hits"])
}

// 处理数据索引
func ExportBlogs(db *gorm.DB) ([]model.Blog, error) {
	var blogs []model.Blog
	err := db.Find(&blogs).Error
	if err != nil {
		return nil, err
	}
	return blogs, nil
}

// func IndexBlogsToElasticsearch(esClient *elastic.Client, blogs []model.Blog) error {
// 	for _, blog := range blogs {
// 		// 查询用户信息以获取作者名称
// 		var user model.User

// 		key := fmt.Sprintf("%d", blog.AuthorID)

// 		res := initialize.RedisClient.Get(context.Background(), key)
// 		if res.Err() == nil {
// 			userJSON := res.Val()
// 			err := json.Unmarshal([]byte(userJSON), &user)
// 			if err != nil {
// 				fmt.Println("Error unmarshaling JSON:", err)
// 			}
// 		} else {
// 			// 从数据库中查询用户信息
// 			if err := initialize.DB.Where("id = ?", blog.AuthorID).First(&user).Error; err != nil {
// 				return fmt.Errorf("failed to find user with ID %d: %w", blog.AuthorID, err)
// 			}
// 			userJson, _ := json.Marshal(user)
// 			// 将用户信息缓存到 Redis
// 			initialize.RedisClient.Set(context.Background(), key, userJson, 0)
// 		}

// 		// if err := initialize.DB.Where("id = ?", blog.AuthorID).First(&user).Error; err != nil {
// 		// 	return fmt.Errorf("failed to find user with ID %d: %w", blog.AuthorID, err)
// 		// }

// 		// 构建 Elasticsearch 文档
// 		esBlog := map[string]interface{}{
// 			"id":        blog.ID,
// 			"title":     blog.Title,
// 			"content":   blog.Content,
// 			"author_id": blog.AuthorID,
// 			"author": map[string]interface{}{
// 				"id":   user.ID,
// 				"name": user.Username, // 假设 User 有一个 Name 字段
// 			},
// 			"published_at": blog.PublishedAt,
// 			"created_at":   blog.CreatedAt,
// 			"updated_at":   blog.UpdatedAt,
// 		}

// 		// 索引文档到 Elasticsearch
// 		_, err := esClient.Index().
// 			Index("blogs").
// 			Id(fmt.Sprintf("%d", blog.ID)).
// 			BodyJson(esBlog).
// 			Refresh("wait_for").
// 			Do(context.Background())
// 		if err != nil {
// 			return fmt.Errorf("failed to index blog to Elasticsearch: %w", err)
// 		}
// 	}
// 	return nil
// }

func EndexBlogToElasticsearch(esBlog map[string]interface{}) {
	id := esBlog["id"].(float64)
	body, _ := json.Marshal(esBlog)

	log.Printf("Indexing blog with ID %d", int(id))

	req := esapi.IndexRequest{
		Index:      "blogs",
		DocumentID: fmt.Sprintf("%d", int(id)),
		Body:       strings.NewReader(string(body)),
		OpType:     "index",
		Refresh:    "wait_for",
	}
	res, err := req.Do(context.Background(), initialize.EsClient)
	if err != nil {
		log.Printf("Failed to index blog to Elasticsearch: %v", err)
		return
	}
	defer res.Body.Close()
	if res.IsError() {
		log.Printf("[%s] Error indexing document: %s", res.Status(), res.String())
	} else {
		log.Printf("Successfully indexed blog with ID %d", int(id))
	}
}

func IndexBlogToElasticsearch(blog model.Blog) error {
	// 查询用户信息以获取作者名称
	var user model.User
	if err := initialize.DB.Where("id = ?", blog.AuthorID).First(&user).Error; err != nil {
		return err
	}

	// 构建 Elasticsearch 文档
	esBlog := map[string]interface{}{
		"id":        blog.ID,
		"title":     blog.Title,
		"content":   blog.Content,
		"author_id": blog.AuthorID,
		"author": map[string]interface{}{
			"id":   user.ID,
			"name": user.Username, // 假设 User 有一个 Name 字段
		},
		"published_at": blog.PublishedAt,
		"created_at":   blog.CreatedAt,
		"updated_at":   blog.UpdatedAt,
	}

	// 将文档序列化为 JSON 字符串
	body, err := json.Marshal(esBlog)
	if err != nil {
		log.Printf("Failed to marshal blog document: %v", err)
		return err
	}

	// 索引文档到 Elasticsearch
	req := esapi.IndexRequest{
		Index:      "blogs",
		DocumentID: fmt.Sprintf("%d", blog.ID), // 使用正确的字段名称 DocumentID
		Body:       strings.NewReader(string(body)),
		OpType:     "index",
		Refresh:    "wait_for", // 立即刷新索引
	}

	// 执行请求
	res, err := req.Do(context.Background(), initialize.EsClient)
	if err != nil {
		log.Printf("Error getting response: %v", err)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		json.NewDecoder(res.Body).Decode(&e)
		log.Printf("Error indexing blog %d: %v", blog.ID, e)
		return fmt.Errorf("error indexing blog %d", blog.ID)
	}

	return nil
}
