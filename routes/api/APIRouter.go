package api

import (
	"context"
	"fmt"
	"net/http"

	"pioneerwebworks.com/juniper/auth"

	"github.com/gorilla/mux"
)

type APIRouter struct {
	Mux        *mux.Router
	Context    context.Context
	AuthRouter http.Handler
	PostRouter http.Handler
}

func NewAPIRouter(context context.Context, parentRouter *mux.Router) *APIRouter {
	a := &APIRouter{
		Mux:     parentRouter.Mux.SubRouter(),
		Context: context,
	}
	a.routes()
	return a
}

func (a *APIRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("APIRouter.ServeHTTP" + r.URL.Path)
	a.mux.ServeHTTP(w, r)
}

func (a *APIRouter) routes() {
	a.AuthRouter = NewAuthRouter(a.Context)
	a.PostRouter = NewPostRouter(a.Context)
	//a.mux.PathPrefix("/auth").Handler(a.AuthRouter)
	//a.mux.HandleFunc("/auth/login", HandleLogin).Methods("POST")
	a.mux.PathPrefix("/post").Handler(auth.WithAuth(a.PostRouter))
}
