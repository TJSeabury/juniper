package models

import (
	"time"
)

type User struct {
	ID            uint      `gorm:"primaryKey"`
	Username      string    `gorm:"size:255;not null"`
	Password      string    `gorm:"size:255;not null"`
	Email         string    `gorm:"size:255;not null;unique"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Forename      string    `gorm:"size:255;not null"`
	Surname       string    `gorm:"size:255;not null"`
	Birthdate     time.Time `gorm:"not null"`
	EmailVerified bool      `gorm:"default:false"`
	PhoneNumber   string    `gorm:"size:255;not null"`
	PhoneVerified bool      `gorm:"default:false"`
	UserRole      string    `gorm:"size:255;not null"`
}
