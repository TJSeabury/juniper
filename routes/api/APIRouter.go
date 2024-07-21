package api

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	store *sessions.CookieStore
)

func init() {
	key, err := loadSessionKey()
	if err != nil {
		key, err = GenerateRandomKey(32)
		if err != nil {
			log.Fatal(err)
		}
		err = saveSessionKey(key)
		if err != nil {
			log.Fatal(err)
		}
	}

	store = sessions.NewCookieStore([]byte(key))
}

type APIRouter struct {
	mux        *http.ServeMux
	Context    context.Context
	AuthRouter http.Handler
	PostRouter http.Handler
}

func NewAPIRouter() *APIRouter {
	a := &APIRouter{mux: http.NewServeMux()}
	a.routes()
	return a
}

func (a *APIRouter) routes() {
	a.AuthRouter = NewAuthRouter()
	a.PostRouter = NewPostRouter()
	a.mux.Handle("/auth/", a.AuthRouter)
	a.mux.Handle("/post/", a.PostRouter)
}

func (a *APIRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}
