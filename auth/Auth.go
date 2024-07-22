package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
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

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	auth, _ := session.Values["authenticated"].(bool)
	userID, _ := session.Values["userID"].(int)

	if !auth {
		if r.Method == "GET" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized"}`))
		return
	}

	ctx := context.WithValue(r.Context(), userIDKey, userID)
	r = r.WithContext(ctx)

	am.Next.ServeHTTP(w, r)
}

func WithAuth(next http.Handler) http.Handler {
	return &AuthMiddleware{Next: next}
}
