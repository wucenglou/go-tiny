package controller

import (
	"go-tiny/initialize"
	"go-tiny/model"
	"go-tiny/utils"

	"github.com/gin-gonic/gin"
)

type BlogController struct{}

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
