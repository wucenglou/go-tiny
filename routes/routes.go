package routes

import (
	"go-tiny/controller"
	"go-tiny/docs"
	"go-tiny/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(r *gin.Engine) {
	userController := controller.UserController{}
	blogController := controller.BlogController{}
	commonController := controller.CommonController{}

	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "I am ok !",
		})
	})

	// docs.SwaggerInfo.Title = "Swagger Example API"
	// docs.SwaggerInfo.Description = "This is a sample server Petstore server."
	// docs.SwaggerInfo.Version = "1.0"
	// docs.SwaggerInfo.Host = "petstore.swagger.io"
	docs.SwaggerInfo.BasePath = "/v2"
	// docs.SwaggerInfo.Schemes = []string{"http", "https"}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 设置通用路由
	commonRouter := r.Group("/api/common")
	{
		// 保护需要认证的路由
		protectedCommonRouter := commonRouter.Group("")
		protectedCommonRouter.Use(middleware.JWTAuth())
		{
			// 添加上传头像的路由
			protectedCommonRouter.POST("/avatar", middleware.UploadMiddleware(), commonController.Upload)

			// 获取当前登录用户信息
			protectedCommonRouter.GET("/user", commonController.GetCurrentUser)
		}
	}

	// 设置用户路由
	userRouter := r.Group("/api/users")
	{
		userRouter.POST("", userController.CreateUser)  // 创建用户
		userRouter.POST("/login", userController.Login) // 登录用户
		userRouter.GET("", userController.GetUsers)     // 获取所有用户列表

		// 保护需要认证的路由
		protectedUserRouter := userRouter.Group("")
		protectedUserRouter.Use(middleware.JWTAuth())
		{
			protectedUserRouter.GET("/user", userController.GetUser)      // 获取单个用户信息
			protectedUserRouter.PUT("/:id", userController.UpdateUser)    // 更新用户
			protectedUserRouter.DELETE("/:id", userController.DeleteUser) // 删除用户
		}
	}

	// 设置博客路由
	blogRouter := r.Group("/api/blogs")
	{
		blogRouter.GET("/:id", blogController.GetBlog) // 获取单个博客
		blogRouter.GET("", blogController.GetBlogs)    // 获取所有博客

		// 保护需要认证的路由
		protectedBlogRouter := blogRouter.Group("")
		protectedBlogRouter.Use(middleware.JWTAuth())
		{
			protectedBlogRouter.POST("", blogController.CreateBlog)
			protectedBlogRouter.PUT("/:id", blogController.UpdateBlog)    // 更新博客
			protectedBlogRouter.DELETE("/:id", blogController.DeleteBlog) // 删除博客
		}
	}
}
