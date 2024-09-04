package controller

import (
	"fmt"
	"go-tiny/initialize"
	"go-tiny/model"
	"go-tiny/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct{}

func (uc *UserController) GetUsers(c *gin.Context) {
	fmt.Println("GetUsers")
	var users []model.User
	result := initialize.DB.Find(&users)
	if result.Error != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(200, users)
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to hash password"})
		return
	}

	user.Password = hashedPassword
	initialize.DB.Create(&user)
	c.JSON(201, user)
}

func (uc *UserController) UpdateUser(c *gin.Context) {
	var user model.User
	if err := initialize.DB.First(&user, c.Param("id")).Error; err != nil {
		c.AbortWithStatusJSON(404, gin.H{"error": "User not found"})
		return
	}

	// 确保只有该用户才能更新自己的资料
	if c.GetString("username") != user.Username {
		c.AbortWithStatusJSON(403, gin.H{"error": "Permission denied"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if user.Password != "" {
		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashedPassword
	}

	initialize.DB.Save(&user)
	c.JSON(200, user)
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	var user model.User
	if initialize.DB.Delete(&user, c.Param("id")).RowsAffected == 0 {
		c.AbortWithStatusJSON(404, gin.H{"error": "User not found"})
		return
	}
	c.JSON(204, gin.H{"message": "User deleted"})
}

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
	result := initialize.DB.Where("username = ?", loginData.Username).First(&user)
	if result.Error != nil {
		c.JSON(404, gin.H{"error": "User not found"})
		return
	}

	if !VerifyPassword(loginData.Password, user.Password) {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

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
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
