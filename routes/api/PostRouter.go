package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/models"
)

type PostRouter struct {
	Mux     *mux.Router
	Context context.Context
}

func NewPostRouter(context context.Context, parentRouter *mux.Router) *PostRouter {
	pr := &PostRouter{
		Mux:     parentRouter.Mux.Subrouter(),
		Context: context,
	}
	pr.routes()
	return pr
}

func (pr *PostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.mux.ServeHTTP(w, r)
}

func (pr *PostRouter) routes() {
	pr.mux.HandleFunc("/post", pr.HandlePost)
}

func (pr *PostRouter) HandlePost(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	w.Header().Set("Accept", "application/json")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")

	type response struct {
		OK      bool        `json:"ok"`
		Message string      `json:"message"`
		Post    models.Post `json:"post"`
	}

	switch r.Method {
	case "GET":
		slug := r.URL.Path[len("/post/"):]
		post := models.Post{}
		post_db.Find(&post, "slug = ?", slug)

		if post.ID == 0 {
			http.NotFound(w, r)
			return
		}

		postJSON, err := json.Marshal(response{
			OK:      true,
			Message: "Post found",
			Post:    post,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(postJSON)
	case "POST":
		r.ParseForm()
		title := r.FormValue("title")
		slug := r.FormValue("slug")
		content := r.FormValue("content")
		post := models.Post{
			Title:   title,
			Slug:    slug,
			Content: content,
		}
		result := post_db.Create(&post)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			responseJSON, err := json.Marshal(response{
				OK:      false,
				Message: result.Error.Error(),
				Post:    models.Post{},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(responseJSON)
			return
		}

		responseJson, err := json.Marshal(response{
			OK:      true,
			Message: "Post created",
			Post:    post,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"ok": false, "message": "Internal server error"}`))
			return
		}

		w.Write(responseJson)
	case "PUT":
		r.ParseForm()
		id := r.FormValue("id")
		slug := r.FormValue("slug")
		title := r.FormValue("title")
		content := r.FormValue("content")
		post := models.Post{}
		post_db.First(&post, id)
		post.Title = title
		post.Slug = slug
		post.Content = content
		result := post_db.Save(&post)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			responseJSON, err := json.Marshal(response{
				OK:      false,
				Message: result.Error.Error(),
				Post:    models.Post{},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(responseJSON)
			return
		}
		responseJson, err := json.Marshal(response{
			OK:      true,
			Message: "Post updated",
			Post:    post,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"ok": false, "message": "Internal server error"}`))
			return
		}
		w.Write(responseJson)

	case "DELETE":
		r.ParseForm()
		id := r.FormValue("id")
		post := models.Post{}
		post_db.First(&post, id)
		result := post_db.Delete(&post)
		if result.Error != nil {
			w.WriteHeader(http.StatusInternalServerError)
			responseJSON, err := json.Marshal(response{
				OK:      false,
				Message: result.Error.Error(),
				Post:    models.Post{},
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(responseJSON)
			return
		}
		responseJson, err := json.Marshal(response{
			OK:      true,
			Message: "Post deleted",
			Post:    post,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"ok": false, "message": "Internal server error"}`))
			return
		}
		w.Write(responseJson)
	}

}
