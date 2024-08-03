package models

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Post struct {
	ID        uint      `gorm:"primaryKey"`
	Slug      string    `gorm:"size:255;not null"`
	Title     string    `gorm:"size:255;not null"`
	Content   string    `gorm:"size:255;not null"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UserID    uint      `gorm:"not null"`
}

type PostModelHandler struct {
	db *gorm.DB
	r  *http.ServeMux
}

func NewPostModelHandler(
	router *http.ServeMux,
	context context.Context,
	allowedOrigins []string,
	allowedMethods []string,
) *PostModelHandler {
	user_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	postHandler := &PostModelHandler{
		user_db,
		router,
	}

	postHandler.RegisterHandlers(context)

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedMethods:   allowedMethods,
	})

	c.Handler(postHandler.r)

	return postHandler
}

func (h *PostModelHandler) Create(u *Post) error {
	tx := h.db.Create(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) Read(u *Post) error {
	tx := h.db.First(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) Update(u *Post) error {
	tx := h.db.Save(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) Delete(u *Post) error {
	tx := h.db.Delete(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) Find(u *Post) error {
	tx := h.db.Find(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) List(criteria Post) ([]Post, error) {
	var posts []Post
	h.db.Find(&posts)
	return posts, nil
}

func (h *PostModelHandler) Count(criteria Post) (int64, error) {
	var count int64
	h.db.Count(&count)
	return count, nil
}

func (h *PostModelHandler) Exists(p *Post) (bool, error) {
	var exists bool
	tx := h.db.Where("slug = ?", p.Slug).First(&exists)
	if tx.Error != nil {
		return exists, tx.Error
	}
	return exists, nil
}

func (h *PostModelHandler) BatchCreate(items []Post) error {
	tx := h.db.Create(&items)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) BatchUpdate(items []Post) error {
	tx := h.db.Save(&items)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) BatchDelete(ids []int) error {
	tx := h.db.Delete(&ids)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *PostModelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func (h *PostModelHandler) RegisterHandlers(context context.Context) {
	h.r.HandleFunc("GET /api/posts/{slug}", h.Handle_Get_One)
	h.r.HandleFunc("GET /api/posts/", h.Handle_Get_List)
	h.r.HandleFunc("POST /api/posts/", h.Handle_Post)
	h.r.HandleFunc("PUT /api/posts/{slug}", h.Handle_Put)
	h.r.HandleFunc("DELETE /api/posts/{slug}", h.Handle_Delete)

	notFoundPatterns := []string{
		"PUT /api/posts/",
		"DELETE /api/posts/",
		"GET /api/posts/{slug}/...",
		"POST /api/posts/{slug}/...",
		"PUT /api/posts/{slug}/...",
		"DELETE /api/posts/{slug}/...",
	}
	for _, pattern := range notFoundPatterns {
		h.r.HandleFunc(pattern, h.Handle_NotFound)
	}

}

func (h *PostModelHandler) Handle_Get_One(
	w http.ResponseWriter,
	r *http.Request,
) {
	slug := r.PathValue("slug")
	post := Post{}
	tx := h.db.First(&post, slug)
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h *PostModelHandler) Handle_NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Endpoint not found."})
}

func (h *PostModelHandler) Handle_Get_List(
	w http.ResponseWriter,
	r *http.Request,
) {
	posts := []Post{}
	h.db.Find(&posts)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(posts)
}

func (h *PostModelHandler) Handle_Post(
	w http.ResponseWriter,
	r *http.Request,
) {
	type postForm struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
	}
	var data postForm
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post := Post{
		Title:   data.Title,
		Slug:    data.Slug,
		Content: data.Content,
	}
	result := h.db.Create(&post)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result.Error.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h *PostModelHandler) Handle_Put(
	w http.ResponseWriter,
	r *http.Request,
) {
	slug := r.PathValue("slug")
	post := Post{}
	h.db.First(&post, slug)
	type postForm struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
	}
	var data postForm
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	post.Title = data.Title
	post.Slug = data.Slug
	post.Content = data.Content
	result := h.db.Save(&post)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result.Error.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

func (h *PostModelHandler) Handle_Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	slug := r.PathValue("slug")
	post := Post{}
	h.db.First(&post, slug)
	result := h.db.Delete(&post)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(result.Error.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

// type PostRouter struct {
// 	Context context.Context
// }

// func (pr *PostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
// 	if err != nil {
// 		panic("failed to connect database")
// 	}

// 	w.Header().Set("Accept", "application/json")
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")

// 	type response struct {
// 		OK      bool        `json:"ok"`
// 		Message string      `json:"message"`
// 		Post    models.Post `json:"post"`
// 	}

// 	switch r.Method {
// 	case "GET":
// 		slug := r.URL.Path[len("/post/"):]
// 		post := models.Post{}
// 		post_db.Find(&post, "slug = ?", slug)

// 		if post.ID == 0 {
// 			http.NotFound(w, r)
// 			return
// 		}

// 		postJSON, err := json.Marshal(response{
// 			OK:      true,
// 			Message: "Post found",
// 			Post:    post,
// 		})
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}

// 		w.Write(postJSON)
// 	case "POST":
// 		r.ParseForm()
// 		title := r.FormValue("title")
// 		slug := r.FormValue("slug")
// 		content := r.FormValue("content")
// 		post := models.Post{
// 			Title:   title,
// 			Slug:    slug,
// 			Content: content,
// 		}
// 		result := post_db.Create(&post)
// 		if result.Error != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			responseJSON, err := json.Marshal(response{
// 				OK:      false,
// 				Message: result.Error.Error(),
// 				Post:    models.Post{},
// 			})
// 			if err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
// 			w.Write(responseJSON)
// 			return
// 		}

// 		responseJson, err := json.Marshal(response{
// 			OK:      true,
// 			Message: "Post created",
// 			Post:    post,
// 		})
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(`{"ok": false, "message": "Internal server error"}`))
// 			return
// 		}

// 		w.Write(responseJson)
// 	case "PUT":
// 		r.ParseForm()
// 		id := r.FormValue("id")
// 		slug := r.FormValue("slug")
// 		title := r.FormValue("title")
// 		content := r.FormValue("content")
// 		post := models.Post{}
// 		post_db.First(&post, id)
// 		post.Title = title
// 		post.Slug = slug
// 		post.Content = content
// 		result := post_db.Save(&post)
// 		if result.Error != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			responseJSON, err := json.Marshal(response{
// 				OK:      false,
// 				Message: result.Error.Error(),
// 				Post:    models.Post{},
// 			})
// 			if err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
// 			w.Write(responseJSON)
// 			return
// 		}
// 		responseJson, err := json.Marshal(response{
// 			OK:      true,
// 			Message: "Post updated",
// 			Post:    post,
// 		})
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(`{"ok": false, "message": "Internal server error"}`))
// 			return
// 		}
// 		w.Write(responseJson)

// 	case "DELETE":
// 		r.ParseForm()
// 		id := r.FormValue("id")
// 		post := models.Post{}
// 		post_db.First(&post, id)
// 		result := post_db.Delete(&post)
// 		if result.Error != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			responseJSON, err := json.Marshal(response{
// 				OK:      false,
// 				Message: result.Error.Error(),
// 				Post:    models.Post{},
// 			})
// 			if err != nil {
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 				return
// 			}
// 			w.Write(responseJSON)
// 			return
// 		}
// 		responseJson, err := json.Marshal(response{
// 			OK:      true,
// 			Message: "Post deleted",
// 			Post:    post,
// 		})
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			w.Write([]byte(`{"ok": false, "message": "Internal server error"}`))
// 			return
// 		}
// 		w.Write(responseJson)
// 	}

// }
