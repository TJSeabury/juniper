package api

import (
	"context"
	"net/http"
)

type AuthRouter struct {
	mux     *http.ServeMux
	Context context.Context
}

func NewAuthRouter() *AuthRouter {
	ar := &AuthRouter{mux: http.NewServeMux()}
	ar.routes()
	return ar
}

func (ar *AuthRouter) routes() {
	ar.mux.HandleFunc("/login", ar.HandleLogin)
	ar.mux.HandleFunc("/logout", ar.HandleLogout)
}

func (ar *AuthRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ar.mux.ServeHTTP(w, r)
}

func (ar *AuthRouter) HandleLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Authentication goes here
	// ...

	// Set user as authenticated
	session.Values["authenticated"] = true
	session.Save(r, w)
}

func (ar *AuthRouter) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
}
