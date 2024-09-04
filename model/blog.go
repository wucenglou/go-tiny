package model

import (
	"time"

	"gorm.io/gorm"
)

type Blog struct {
	gorm.Model
	Title       string    `gorm:"not null;index;unique" json:"title"`
	Content     string    `gorm:"not null" json:"content"`
	AuthorID    uint      `gorm:"not null" json:"-"`
	Author      User      `gorm:"foreignKey:AuthorID" json:"author"`
	PublishedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"published_at"`
}
