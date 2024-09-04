package controller

import (
	"go-tiny/initialize"
	"go-tiny/model"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type CommonController struct{}

// Upload handles the file upload request.
func (cc *CommonController) Upload(c *gin.Context) {
	// 从上下文中获取文件名
	filename, exists := c.Get("filename")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file was uploaded"})
		return
	}

	// 将接口类型转换为字符串
	avatarFilename := filename.(string)

	// 假设您有一个实体ID从JWT或其他来源获取
	// 为了简单起见，我们假设它是通过上下文传递的
	entityId, exists := c.Get("entityId")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing entity ID"})
		return
	}

	entityID := entityId.(uint)

	// 更新数据库中实体的头像路径
	var user model.User
	if err := initialize.DB.Where("id = ?", entityID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 假设头像URL是相对于uploads目录的
	user.Avatar = filepath.Join("/uploads/", avatarFilename)

	if err := initialize.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar uploaded successfully", "avatar_url": user.Avatar})
}

// GetCurrentUser returns the current authenticated user.
func (cc *CommonController) GetCurrentUser(c *gin.Context) {
	// 假设您有一个当前用户的ID在上下文中，来自JWT中间件
	currentUserId, exists := c.Get("currentUserId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id, ok := currentUserId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	var user model.User
	if err := initialize.DB.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
