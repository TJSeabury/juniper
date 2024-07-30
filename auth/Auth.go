package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"pioneerwebworks.com/juniper/models"
)

var (
	Store *sessions.CookieStore
)

func Init() {
	key, err := LoadSessionKey()
	if err != nil {
		key, err = GenerateRandomKey(32)
		if err != nil {
			log.Fatal(err)
		}
		err = SaveSessionKey(key)
		if err != nil {
			log.Fatal(err)
		}
	}

	Store = sessions.NewCookieStore([]byte(key))
}

type contextKey string

const userIDKey contextKey = "userID"

func GenerateRandomKey(length int) ([]byte, error) {
	if length != 16 && length != 24 && length != 32 {
		return nil, fmt.Errorf("invalid key length: %d", length)
	}

	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func SaveSessionKey(key []byte) error {
	err := os.WriteFile("session.key", key, 0600)
	if err != nil {
		return err
	}
	return nil
}

func LoadSessionKey() ([]byte, error) {
	key, err := os.ReadFile("session.key")
	if err != nil {
		return nil, err
	}
	return key, nil
}

type AuthMiddleware struct {
	Next http.Handler
}

func (am *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "juniper-session")
	userDB := models.ConnectToUserDB()

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	auth, authOK := session.Values["authenticated"].(bool)
	userID, _ := session.Values["userID"].(uint)

	user, err := userDB.GetUser(uint(userID))

	log.Println("user", user)

	if err != nil {
		log.Println("AUTH: Error getting user: " + err.Error())
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Println("AUTH: Auth: " + strconv.FormatBool(auth))
	log.Println("AUTH: Auth OK: " + strconv.FormatBool(authOK))
	log.Println("AUTH: User ID: " + strconv.Itoa(int(userID)))

	if !auth {
		log.Println("AUTH: Not authenticated")
		if r.Method == "GET" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	if !user.EmailVerified {
		log.Println("AUTH: Email not verified")
		if r.Method == "GET" {
			http.Redirect(w, r, "/verify", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized. Please verify your email address."}`))
		return
	}

	ctx := context.WithValue(r.Context(), userIDKey, userID)
	r = r.WithContext(ctx)

	am.Next.ServeHTTP(w, r)
}

func WithAuth(next http.Handler) http.Handler {
	return &AuthMiddleware{Next: next}
}

// GenerateToken generates a secure, unique token for email verification.
func GenerateToken(email string) (string, error) {
	// Generate a UUID.
	uuidWithHyphen, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	uuid := uuidWithHyphen.String()

	// Optionally, combine it with user-specific information.
	combined := uuid + ":" + email

	// Hash the combined string.
	hasher := sha256.New()
	_, err = hasher.Write([]byte(combined))
	if err != nil {
		return "", err
	}
	hashed := hasher.Sum(nil)

	// Encode the hash in base64.
	token := base64.URLEncoding.EncodeToString(hashed)

	return token, nil
}

// GenerateSecureRandomToken generates a secure random token using crypto/rand.
func GenerateSecureRandomToken() (string, error) {
	b := make([]byte, 32) // Generates a 256-bit token.
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	// Encode the binary data to a base64 URL encoded string.
	token := base64.URLEncoding.EncodeToString(b)

	return token, nil
}
