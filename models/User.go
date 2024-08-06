package models

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID            uint      `gorm:"primarykey"`
	CreatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt     time.Time `json:"deletedAt"`
	Username      string    `gorm:"size:255;not null" json:"username"`
	Password      string    `gorm:"size:255;not null" json:"password"`
	Email         string    `gorm:"size:255;not null;unique" json:"email"`
	LastLoginAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"lastLoginAt"`
	Forename      string    `gorm:"size:255;not null" json:"forename"`
	Surname       string    `gorm:"size:255;not null" json:"surname"`
	Birthdate     time.Time `gorm:"not null" json:"birthdate"`
	EmailToken    string    `gorm:"size:255" json:"emailToken"`
	EmailVerified bool      `gorm:"default:false" json:"emailVerified"`
	PhoneNumber   string    `gorm:"size:255;not null" json:"phoneNumber"`
	PhoneVerified bool      `gorm:"default:false" json:"phoneVerified"`
	UserRole      string    `gorm:"size:255;not null" json:"userRole"`
}

func UserJSONMapper(data map[string]interface{}) (User, error) {
	parsedID, IDOK := data["id"].(float64)
	parsedUsername, UsernameOK := data["username"].(string)
	parsedPassword, PasswordOK := data["password"].(string)
	parsedEmail, EmailOK := data["email"].(string)
	parsedLastLoginAt, LastLoginAtOK := data["lastLoginAt"].(time.Time)
	parsedForename, ForenameOK := data["forename"].(string)
	parsedSurname, SurnameOK := data["surname"].(string)
	parsedBirthdate, BirthdateOK := data["birthdate"].(time.Time)
	parsedEmailToken, EmailTokenOK := data["emailToken"].(string)
	parsedEmailVerified, EmailVerifiedOK := data["emailVerified"].(bool)
	parsedPhoneNumber, PhoneNumberOK := data["phoneNumber"].(string)
	parsedPhoneVerified, PhoneVerifiedOK := data["phoneVerified"].(bool)
	parsedUserRole, UserRoleOK := data["userRole"].(string)

	oks := []bool{IDOK, UsernameOK, PasswordOK, EmailOK, LastLoginAtOK, ForenameOK, SurnameOK, BirthdateOK, EmailTokenOK, EmailVerifiedOK, PhoneNumberOK, PhoneVerifiedOK, UserRoleOK}

	if !IDOK || !UsernameOK || !PasswordOK || !EmailOK || !LastLoginAtOK || !ForenameOK || !SurnameOK || !BirthdateOK || !EmailTokenOK || !EmailVerifiedOK || !PhoneNumberOK || !PhoneVerifiedOK || !UserRoleOK {
		fmt.Printf("invalid data: %+v\n", oks)
		return User{}, errors.New("invalid data")
	}

	return User{
		ID:            uint(parsedID),
		Username:      parsedUsername,
		Password:      parsedPassword,
		Email:         parsedEmail,
		LastLoginAt:   parsedLastLoginAt,
		Forename:      parsedForename,
		Surname:       parsedSurname,
		Birthdate:     parsedBirthdate,
		EmailToken:    parsedEmailToken,
		EmailVerified: parsedEmailVerified,
		PhoneNumber:   parsedPhoneNumber,
		PhoneVerified: parsedPhoneVerified,
		UserRole:      parsedUserRole,
	}, nil
}

type UserDB struct {
	DB *gorm.DB
}

func ConnectToUserDB() UserDB {
	user_db, err := gorm.Open(sqlite.Open("database/user.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return UserDB{DB: user_db}
}

func (udb *UserDB) CreateUser(u *User) (uint, error) {
	db := udb.DB
	tx := db.Create(&u)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return u.ID, nil
}

func (udb *UserDB) GetUser(id uint) (User, error) {
	db := udb.DB

	var u User
	db.First(&u, id)

	if u.ID == 0 {
		return u, errors.New("user not found")
	}

	return u, nil
}

func (udb *UserDB) UpdateUser(u User) {
	db := udb.DB
	db.Save(&u)
}

func (udb *UserDB) DeleteUser(id string) {
	db := udb.DB

	var u User
	db.First(&u, id)
	db.Delete(&u)
}

func (udb *UserDB) FindByUsername(username string) User {
	var u User
	udb.DB.Where("username = ?", username).First(&u)
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

func (u *User) CheckPassword(password string) bool {
	hashedPasswordInDB := u.Password

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPasswordInDB),
		[]byte(password),
	) == nil
}

func HashEmailToken(email string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword(
		[]byte(email),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (u *User) CheckEmailToken(token string) bool {
	hashedEmailTokenInDB := u.EmailToken

	return bcrypt.CompareHashAndPassword(
		[]byte(hashedEmailTokenInDB),
		[]byte(token),
	) == nil
}
