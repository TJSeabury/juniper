package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID            uint      `gorm:"primaryKey"`
	Username      string    `gorm:"size:255;not null"`
	Password      string    `gorm:"size:255;not null"`
	Email         string    `gorm:"size:255;not null;unique"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	LastLoginAt   time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Forename      string    `gorm:"size:255;not null"`
	Surname       string    `gorm:"size:255;not null"`
	Birthdate     time.Time `gorm:"not null"`
	EmailVerified bool      `gorm:"default:false"`
	PhoneNumber   string    `gorm:"size:255;not null"`
	PhoneVerified bool      `gorm:"default:false"`
	UserRole      string    `gorm:"size:255;not null"`
}

func (u *User) CheckPassword(password string) bool {
	hashedPasswordInDB := u.Password

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPasswordInDB),
		[]byte(password),
	) == nil
}

type UserDB struct {
	db *gorm.DB
}

func ConnectToUserDB() UserDB {
	user_db, err := gorm.Open(sqlite.Open("database/user.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return UserDB{db: user_db}
}

func (udb *UserDB) CreateUser(u User) {
	db := udb.db
	db.Create(&u)
}

func (udb *UserDB) GetUser(id string) User {
	db := udb.db

	var u User
	db.First(&u, id)
	return u
}

func (udb *UserDB) UpdateUser(u User) {
	db := udb.db
	db.Save(&u)
}

func (udb *UserDB) DeleteUser(id string) {
	db := udb.db

	var u User
	db.First(&u, id)
	db.Delete(&u)
}

func (udb *UserDB) FindByUsername(username string) User {
	var u User
	udb.db.Where("username = ?", username).First(&u)
	return u
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
