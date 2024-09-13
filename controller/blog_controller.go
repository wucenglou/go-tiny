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

	"github.com/gin-gonic/gin"
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

	// 构建 Elasticsearch 查询语句
	queryJSON := fmt.Sprintf(`
	{
		"query": {
			"multi_match": {
				"query": "%s",
				"fields": ["title^2", "content"],
				"fuzziness": "AUTO"
			}
		}
	}`, query)

	// 执行 Elasticsearch 查询
	searchResult, err := initialize.EsClient.Search().
		Index("blogs").
		BodyString(queryJSON).
		Do(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to search blogs"})
		return
	}

	// 解析查询结果
	blogs := make([]model.Blog, 0, len(searchResult.Hits.Hits))
	for _, hit := range searchResult.Hits.Hits {
		var blog model.Blog
		err := json.Unmarshal(hit.Source, &blog)
		if err != nil {
			log.Printf("Error unmarshaling hit source: %v", err)
			continue
		}
		blogs = append(blogs, blog)
	}

	// 返回查询结果
	c.JSON(200, blogs)
}
