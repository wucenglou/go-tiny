package routes

import (
	"go-tiny/controller"
	"go-tiny/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	userController := controller.UserController{}
	blogController := controller.BlogController{}

	// 设置用户路由
	userRouter := r.Group("/api/users")
	{
		userRouter.POST("", userController.CreateUser)
		userRouter.POST("/login", userController.Login)
		userRouter.GET("", userController.GetUsers) // 保持没有尾部斜杠

		// 保护需要认证的路由
		protectedUserRouter := userRouter.Group("") // 使用空字符串来避免尾部斜杠
		protectedUserRouter.Use(middleware.JWTAuth())
		{
			protectedUserRouter.PUT("/:id", userController.UpdateUser)
			protectedUserRouter.DELETE("/:id", userController.DeleteUser)
		}
	}

	// 设置博客路由
	blogRouter := r.Group("/api/blogs")
	{
		blogRouter.POST("", blogController.CreateBlog)

		// 保护需要认证的路由
		protectedBlogRouter := blogRouter.Group("") // 使用空字符串来避免尾部斜杠
		protectedBlogRouter.Use(middleware.JWTAuth())
		{
			protectedBlogRouter.PUT("/:id", blogController.UpdateBlog)
			protectedBlogRouter.DELETE("/:id", blogController.DeleteBlog)
			protectedBlogRouter.GET("/:id", blogController.GetBlog)
			protectedBlogRouter.GET("", blogController.GetBlogs) // 保持没有尾部斜杠
		}
	}
}
