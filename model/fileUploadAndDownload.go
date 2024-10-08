package model

import (
	"time"

	"gorm.io/gorm"
)

type FileUploadAndDownload struct {
	gorm.Model
	Name      string         `json:"name" gorm:"comment:文件名"` // 文件名
	Url       string         `json:"url" gorm:"comment:文件地址"` // 文件地址
	Tag       string         `json:"tag" gorm:"comment:文件标签"` // 文件标签
	Key       string         `json:"key" gorm:"comment:编号"`   // 编号
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}
