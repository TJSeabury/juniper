package models

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func Test_Hashpassword(t *testing.T) {
	password := "password"
	hashedPassword, err := HashPassword(password) // Assuming HashPassword uses bcrypt
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
	}
	if hashedPassword == "" {
		t.Errorf("Hashed password is empty")
	}

	// Compare the hashed password with the original password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		t.Errorf("Failed to verify hashed password: %v", err)
	}

	// Attempt to hash the password again to see if we get a different result
	hashedPassword2, err := HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password again: %v", err)
	}
	if hashedPassword == hashedPassword2 {
		t.Errorf("Hashed passwords are the same")
	}

	// Compare the second hashed password with the original password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword2), []byte(password))
	if err != nil {
		t.Errorf("Failed to verify second hashed password: %v", err)
	}
}
