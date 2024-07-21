package routes

import (
	"context"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/views/dashboard"
	"pioneerwebworks.com/juniper/views/public"
)

type DashboardRouter struct {
	mux     *http.ServeMux
	Context context.Context
}

func NewDashboardRouter() *DashboardRouter {
	dr := &DashboardRouter{mux: http.NewServeMux()}
	dr.routes()
	return dr
}

func (dr *DashboardRouter) routes() {
	dr.mux.HandleFunc("/dashboard", dr.HandleDashboard)
}

func (dr *DashboardRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dr.mux.ServeHTTP(w, r)
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
