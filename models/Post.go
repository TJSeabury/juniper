package models

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
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
}

func NewPostModelHandler() *PostModelHandler {
	user_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return &PostModelHandler{
		user_db,
	}
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

func (h *PostModelHandler) RegisterHandlers(r *mux.Router) {
	r.HandleFunc("/api/post/{slug}", h.Handle_Get_One).Methods("GET")
	r.HandleFunc("/api/posts/", h.Handle_Get_List).Methods("GET")
	r.HandleFunc("/api/post/{slug}", h.Handle_Post).Methods("POST")
	r.HandleFunc("/api/post/{slug}", h.Handle_Put).Methods("PUT")
	r.HandleFunc("/api/post/{slug}", h.Handle_Delete).Methods("DELETE")
}

func (h *PostModelHandler) Handle_Get_One(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	slug := vars["slug"]
	post := Post{}
	h.db.First(&post, slug)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
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
	vars := mux.Vars(r)
	slug := vars["slug"]
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
	vars := mux.Vars(r)
	slug := vars["slug"]
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
