package controller

import (
	"fmt"
	"go-tiny/initialize"
	"go-tiny/model"
	"go-tiny/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct{}

// GetUsers retrieves a list of all users.
func (uc *UserController) GetUsers(c *gin.Context) {
	var users []model.User
	result := initialize.DB.Find(&users)
	if result.Error != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(200, users)
}

// CreateUser creates a new user with a hashed password.
func (uc *UserController) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		// 如果绑定请求体到 user 结构体失败，则返回错误
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	if initialize.DB.Where("username = ?", user.Username).First(&user).RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// 对密码进行哈希处理
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = hashedPassword

	// 创建新用户
	initialize.DB.Create(&user)
	// 构建响应结构体
	responseUser := model.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		// 其他需要返回的字段
	}

	c.JSON(201, responseUser)
}

// UpdateUser updates an existing user's information.
func (uc *UserController) UpdateUser(c *gin.Context) {
	var user model.User
	if err := initialize.DB.First(&user, c.Param("id")).Error; err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "User not found"})
		return
	}

	// 检查是否是当前用户本人更新信息
	if c.GetString("username") != user.Username {
		c.AbortWithStatusJSON(403, gin.H{"error": "Permission denied"})
		return
	}

	// 绑定请求体到 user 结构体
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 如果密码字段有值，则对其进行哈希处理
	if user.Password != "" {
		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}

	// 保存更新后的用户信息
	initialize.DB.Save(&user)
	c.JSON(200, user)
}

// DeleteUser deletes an existing user.
func (uc *UserController) DeleteUser(c *gin.Context) {
	var user model.User
	if initialize.DB.Delete(&user, c.Param("id")).RowsAffected == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "User not found"})
		return
	}
	c.JSON(204, gin.H{"message": "User deleted"})
}

// Login authenticates a user and generates a JWT token.
func (uc *UserController) Login(c *gin.Context) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	err := initialize.DB.Where("username = ?", loginData.Username).First(&user).Error
	if err != nil {
		fmt.Println("Database Error:", err)
		c.JSON(401, gin.H{"error": "user not found"})
		return
	}

	// 验证密码
	match, err := VerifyPassword(loginData.Password, user.Password)
	if !match || err != nil {
		fmt.Println("Password Verification Error:", err)
		c.JSON(401, gin.H{"error": "invalid password"})
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"token": token})
}

// HashPassword hashes the provided password using bcrypt.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if the provided password matches the stored hashed password.
func VerifyPassword(password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, err
	}
	return true, nil
}
