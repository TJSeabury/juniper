package models

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint      `gorm:"primaryKey"`
	Slug      string    `gorm:"size:255;not null"`
	Title     string    `gorm:"size:255;not null"`
	Content   string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UserID    uint      `gorm:"not null"`
}

func CreatePost(title, content string) {
	db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post := Post{Title: title, Content: content}
	db.Create(&post)
}

func GetPost(id uint) Post {
	db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post := Post{}
	db.First(&post, id)
	return post
}

func GetPosts() []Post {
	db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	posts := []Post{}
	db.Find(&posts)
	return posts
}

func UpdatePost(id uint, title, content string) {
	db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post := Post{}
	db.First(&post, id)
	post.Title = title
	post.Content = content
	db.Save(&post)
}

func DeletePost(id uint) {
	db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post := Post{}
	db.First(&post, id)
	db.Delete(&post)
}
