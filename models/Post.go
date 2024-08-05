package models

import "time"

type Post struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt time.Time `gorm:"index" json:"deletedAt"`
	Slug      string    `gorm:"size:255;not null" json:"slug"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Content   string    `gorm:"size:255;not null" json:"content"`
	UserID    uint      `gorm:"not null" json:"userID"`
}
