package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
)

const csrfKey contextKey = "csrf_token"

type CSRFMiddleware struct {
	Next http.Handler
}

func WithCSRF(next http.Handler) http.Handler {
	return &CSRFMiddleware{Next: next}
}

func (c *CSRFMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	session, _ := Store.Get(r, "juniper-session")

	if r.Method == "GET" {
		// Check if the CSRF token is present in the session.
		// If it is not present, generate a new CSRF token and store it in the session.
		if _, ok := session.Values[csrfKey].(string); !ok {
			csrfToken, err := GenerateCSRFToken(32)
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), csrfKey, csrfToken)
			r = r.WithContext(ctx)
			session.Values[csrfKey] = csrfToken
			session.Save(r, w)
			c.Next.ServeHTTP(w, r)
			return
		}
	}

	if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" {
		// Check if the CSRF token is present in the session.
		// If it is present, validate the CSRF token.
		// If it is not present, deny the request.
		csrfToken, ok := session.Values[csrfKey].(string)
		var csrfTokenInRequest string
		if ok {
			// If the request is a JSON request, validate the CSRF token.
			if r.Header.Get("Content-Type") == "application/json" {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Server error", http.StatusInternalServerError)
					return
				}

				// Ensure the body reader is replaced so it can be consumed again
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				var data map[string]interface{}
				err = json.Unmarshal(bodyBytes, &data)
				if err != nil {
					http.Error(w, "Server error", http.StatusInternalServerError)
					return
				}

				csrfTokenInRequest = data["csrf"].(string)
			} else {
				csrfTokenInRequest = r.FormValue("csrf")
			}
			if csrfTokenInRequest != csrfToken {
				http.Error(w, "CSRF token mismatch", http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

	}
}

func GenerateCSRFToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)
	return token, nil
}

func GetCSRFToken(r *http.Request) string {
	session, _ := Store.Get(r, "juniper-session")
	csrfToken, ok := session.Values[csrfKey].(string)
	if !ok {
		return ""
	}
	return csrfToken
}
