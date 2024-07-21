package routes

import (
	"context"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/views/public"
)

type PublicRouter struct {
	mux     *http.ServeMux
	Context context.Context
}

func NewPublicRouter() *PublicRouter {
	pr := &PublicRouter{mux: http.NewServeMux()}
	pr.routes()
	return pr
}

func (pr *PublicRouter) routes() {
	pr.mux.HandleFunc("/", pr.HandleRoot)
	pr.mux.HandleFunc("/about", pr.HandleAbout)
	pr.mux.HandleFunc("/blog/", pr.HandleBlog)
}

func (pr *PublicRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.mux.ServeHTTP(w, r)
}

func (pr *PublicRouter) HandleRoot(w http.ResponseWriter, r *http.Request) {
	c := public.Paragraph("Home page content.")
	public.App(
		c,
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(pr.Context, w)
}

func (pr *PublicRouter) HandleAbout(w http.ResponseWriter, r *http.Request) {
	public.App(
		public.Page_About(),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(pr.Context, w)
}

func (pr *PublicRouter) HandleBlog(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//Get the post ID from the URL path
	postID := r.URL.Path[len("/blog/"):]
	post := models.Post{}
	post_db.First(&post, postID)
	public.App(
		public.Post(post),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(pr.Context, w)
}
