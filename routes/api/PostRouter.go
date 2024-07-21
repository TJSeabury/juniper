package api

import (
	"context"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/models"
)

type PostRouter struct {
	mux     *http.ServeMux
	Context context.Context
}

func NewPostRouter() *PostRouter {
	pr := &PostRouter{mux: http.NewServeMux()}
	pr.routes()
	return pr
}

func (pr *PostRouter) routes() {
	pr.mux.HandleFunc("/post", pr.HandlePost)
}

func (pr *PostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.mux.ServeHTTP(w, r)
}

func (pr *PostRouter) HandlePost(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	if r.Method == "POST" {
		r.ParseForm()
		title := r.FormValue("title")
		content := r.FormValue("content")
		post := models.Post{Title: title, Content: content}
		post_db.Create(&post)
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (pr *PostRouter) createPost(title, content string) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post := models.Post{Title: title, Content: content}
	post_db.Create(&post)
}
