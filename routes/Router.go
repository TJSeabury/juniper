package routes

import (
	"context"
	"net/http"

	"pioneerwebworks.com/juniper/routes/api"
)

type Router struct {
	mux             *http.ServeMux
	Context         context.Context
	APIRouter       http.Handler
	DashboardRouter http.Handler
	PublicRouter    http.Handler
}

func NewRouter() *Router {
	r := &Router{mux: http.NewServeMux()}
	r.routes()
	return r
}

func (r *Router) routes() {
	r.DashboardRouter = NewDashboardRouter()
	r.PublicRouter = NewPublicRouter()
	r.APIRouter = api.NewAPIRouter()
	r.mux.Handle("/api/", r.APIRouter)
	r.mux.HandleFunc("/dashboard/", r.DashboardRouter.ServeHTTP)
	r.mux.HandleFunc("/", r.PublicRouter.ServeHTTP)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
