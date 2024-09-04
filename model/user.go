package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `json:"userName" gorm:"comment:用户登录名"` // 用户登录名
	Email    string `json:"email"`
	Password string `json:"-"`
}
