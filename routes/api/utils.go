package api

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
)

// GenerateRandomKey generates a random key of the specified length (16, 24, or 32 bytes).
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

func saveSessionKey(key []byte) error {
	err := os.WriteFile("session.key", key, 0600)
	if err != nil {
		return err
	}
	return nil
}

func loadSessionKey() ([]byte, error) {
	key, err := os.ReadFile("session.key")
	if err != nil {
		return nil, err
	}
	return key, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "cookie-name")

		// Check if user is authenticated
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Call the next handler if authenticated
		next.ServeHTTP(w, r)
	})
}
