package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"userName" gorm:"comment:用户登录名"` // 用户登录名
	Email    string `json:"email"`
	Password string `json:"-"`                                                                   // 密码不返回给客户端
	Avatar   string `json:"avatar" gorm:"type:varchar(255);not null;default:'';comment:用户头像URL"` // 用户头像URL
}
