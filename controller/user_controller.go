package controller

import (
	"fmt"
	"go-tiny/initialize"
	"go-tiny/model"
	"go-tiny/model/common/response"
	"go-tiny/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct{}

// GetUsers retrieves a list of all users.
func (uc *UserController) GetUser(c *gin.Context) {
	// 根据jwt解析用户
	userInfo, ok := c.Get("claims")
	if !ok {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	var user model.User
	userID := userInfo.(*utils.Claims).Id
	err := initialize.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch user"})
		return
	}
	response.OkWithData(gin.H{"user": user}, c)
}

func (uc *UserController) GetUsers(c *gin.Context) {
	// 根据jwt解析用户
	claims, _ := utils.ParseToken(c.GetHeader("Authorization"))
	fmt.Println("claims:", claims)
	var users []model.User
	result := initialize.DB.Find(&users)
	if result.Error != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(200, users)
}

// @Summary 用户注册
// @Description 用户通过用户名和密码注册
// @Tags 用户
// @Accept json
// @Produce json
// @Param userName body string true "用户名"
// @Param email body string true "邮箱"
// @Param password body string true "密码"
// @Success 200 {object} response.Response{data=[]interface{},msg=string} "登录成功，返回用户信息"
// @Router /api/users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	var userReq model.UserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		// 如果绑定请求体到 user 结构体失败，则返回错误
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 检查密码长度
	if len(userReq.Password) < 6 {
		c.JSON(400, gin.H{"error": "Password must be at least 6 characters long"})
		return
	}

	// 检查用户名是否已存在
	user := model.User{}
	if initialize.DB.Where("username = ?", userReq.Username).First(&user).RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	user.Username = userReq.Username
	user.Email = userReq.Email
	user.Password = utils.BcryptHash(userReq.Password)

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
		hashedPassword := utils.BcryptHash(user.Password)
		user.Password = hashedPassword
	} else {
		c.JSON(400, gin.H{"error": "Password cannot be empty"})
		return
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

// @Summary 用户登录
// @Description 用户通过用户名和密码登录系统
// @Tags 用户
// @Accept json
// @Produce json
// @Param username body string true "用户名"
// @Param password body string true "密码"
// @Success 200 {object} response.Response{data=[]interface{},msg=string} "登录成功，返回JWT令牌"
// @Router /api/login [post]
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
	ok := utils.BcryptCheck(loginData.Password, user.Password)
	if !ok {
		c.JSON(401, gin.H{"error": "invalid password"})
		return
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}
	response.OkWithData(gin.H{"token": token}, c)
}
