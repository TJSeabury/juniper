package models

import (
	"errors"
	"fmt"
	"time"
)

type Post struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
	Slug      string    `gorm:"size:255;not null" json:"slug"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Content   string    `gorm:"size:255;not null" json:"content"`
	UserID    uint      `gorm:"not null" json:"userID"`
}

func PostJSONMapper(data map[string]interface{}) (Post, error) {
	parsedID, IDOK := data["id"].(float64)
	parsedTitle, TitleOK := data["title"].(string)
	parsedSlug, SlugOK := data["slug"].(string)
	parsedContent, ContentOK := data["content"].(string)
	parsedUserID, UserIDOK := data["userID"].(float64)

	oks := []bool{IDOK, TitleOK, SlugOK, ContentOK, UserIDOK}

	if !IDOK || !TitleOK || !SlugOK || !ContentOK || !UserIDOK {
		fmt.Printf("invalid data: %+v\n", oks)
		return Post{}, errors.New("invalid data")
	}

	return Post{
		ID:      uint(parsedID),
		Title:   parsedTitle,
		Slug:    parsedSlug,
		Content: parsedContent,
		UserID:  uint(parsedUserID),
	}, nil
}
