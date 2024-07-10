package models

import (
	"time"
)

type Post struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"size:255;not null"`
	Content   string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UserID    uint      `gorm:"not null"`
}
