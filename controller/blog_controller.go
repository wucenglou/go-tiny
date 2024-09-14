package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"go-tiny/initialize"
	"go-tiny/model"
	"go-tiny/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
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

	// 将博客条目索引到 Elasticsearch
	esBlog := map[string]interface{}{
		"id":         blog.ID,
		"title":      blog.Title,
		"content":    blog.Content,
		"author_id":  blog.AuthorID,
		"created_at": blog.CreatedAt,
		"updated_at": blog.UpdatedAt,
	}

	// 索引文档到 Elasticsearch
	_, err := initialize.EsClient.Index().
		Index("blogs").
		Id(fmt.Sprintf("%d", blog.ID)).
		BodyJson(esBlog).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to index blog to Elasticsearch"})
		return
	}

	c.JSON(201, blog)
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
	c.JSON(200, blog)
}

func (bc *BlogController) DeleteBlog(c *gin.Context) {
	var blog model.Blog
	if initialize.DB.Where("id = ?", c.Param("id")).Delete(&blog).RowsAffected == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "Blog not found"})
		return
	}
	c.JSON(204, gin.H{"message": "Blog deleted"})
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
	exists, err := initialize.EsClient.IndexExists("blogs").Do(context.Background())
	if err != nil {
		fmt.Println("Error checking index existence:", err)
	}
	if !exists {
		fmt.Println("Index does not exist")
	}

	// 构建 Elasticsearch 查询
	searchSource := elastic.NewSearchSource()
	searchSource.Query(elastic.NewMultiMatchQuery(query, "title", "content").Type("best_fields").Operator("or").Analyzer("ik_smart"))

	searchResult, err := initialize.EsClient.Search().
		Index("blogs").
		SearchSource(searchSource).
		Do(context.Background())
	if err != nil {
		fmt.Println("Error searching:", err)
	}
	// 返回查询结果
	c.JSON(200, searchResult)
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

func IndexBlogsToElasticsearch(esClient *elastic.Client, blogs []model.Blog) error {
	for _, blog := range blogs {
		// 查询用户信息以获取作者名称
		var user model.User

		key := fmt.Sprintf("%d", blog.AuthorID)

		res := initialize.RedisClient.Get(context.Background(), key)
		if res.Err() == nil {
			userJSON := res.Val()
			err := json.Unmarshal([]byte(userJSON), &user)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
			}
		} else {
			// 从数据库中查询用户信息
			if err := initialize.DB.Where("id = ?", blog.AuthorID).First(&user).Error; err != nil {
				return fmt.Errorf("failed to find user with ID %d: %w", blog.AuthorID, err)
			}
			userJson, _ := json.Marshal(user)
			// 将用户信息缓存到 Redis
			initialize.RedisClient.Set(context.Background(), key, userJson, 0)
		}

		// if err := initialize.DB.Where("id = ?", blog.AuthorID).First(&user).Error; err != nil {
		// 	return fmt.Errorf("failed to find user with ID %d: %w", blog.AuthorID, err)
		// }

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

		// 索引文档到 Elasticsearch
		_, err := esClient.Index().
			Index("blogs").
			Id(fmt.Sprintf("%d", blog.ID)).
			BodyJson(esBlog).
			Refresh("wait_for").
			Do(context.Background())
		if err != nil {
			return fmt.Errorf("failed to index blog to Elasticsearch: %w", err)
		}
	}
	return nil
}

func IndexBlogToElasticsearch(esClient *elastic.Client, blog model.Blog) error {
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

	// 索引文档到 Elasticsearch
	_, err := esClient.Index().
		Index("blogs").
		Id(fmt.Sprintf("%d", blog.ID)).
		BodyJson(esBlog).
		Refresh("wait_for").
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}
