package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"comment:用户登录名"` // 用户登录名
	Email    string `json:"email"`
	Password string `json:"-"  gorm:"comment:用户登录密码"`                                            // 密码不返回给客户端
	Avatar   string `json:"avatar" gorm:"type:varchar(255);not null;default:'';comment:用户头像URL"` // 用户头像URL
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Email    string `json:"email"`
	// 其他需要返回的字段
}
