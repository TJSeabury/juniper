package routes

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/views/dashboard"
	"pioneerwebworks.com/juniper/views/public"
)

type DashboardRouter struct {
	mux     *mux.Router
	Context context.Context
}

func NewDashboardRouter(context context.Context) *DashboardRouter {
	dr := &DashboardRouter{
		mux:     mux.NewRouter(),
		Context: context,
	}
	dr.routes()
	return dr
}

func (dr *DashboardRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dr.mux.ServeHTTP(w, r)
}

func (dr *DashboardRouter) routes() {
	dr.mux.PathPrefix("/dashboard").HandlerFunc(dr.HandleDashboard)
}

func (dr *DashboardRouter) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	posts := []models.Post{}
	post_db.Find(&posts)
	public.App(
		dashboard.Dashboard(posts),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(dr.Context, w)
}
