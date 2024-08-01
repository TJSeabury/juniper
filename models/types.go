package models

import (
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Creater[T any] interface {
	Create(T) error
}

type Reader[T any] interface {
	Read(T) error
}

type Updater[T any] interface {
	Update(T) error
}

type Deleter[T any] interface {
	Delete(T) error
}

type Finder[T any] interface {
	Find(T) error
}

type UniqueFinder[T any] interface {
	FindUnique(T) error
}

type Lister[T any] interface {
	List(criteria T) ([]T, error)
}

type Counter[T any] interface {
	Count(criteria T) (int, error)
}

type ExistChecker[T any] interface {
	Exists(criteria T) (bool, error)
}

type BatchCreater[T any] interface {
	BatchCreate(items []T) error
}

type BatchUpdater[T any] interface {
	BatchUpdate(items []T) error
}

type BatchDeleter[T any] interface {
	BatchDelete(ids []int) error
}

type Transactioner interface {
	Begin() error
	Commit() error
	Rollback() error
}

type Modeller[T any] interface {
	Creater[T]
	Reader[T]
	Updater[T]
	Deleter[T]
	Finder[T]
	UniqueFinder[T]
	Lister[T]
	Counter[T]
	ExistChecker[T]
	BatchCreater[T]
	BatchUpdater[T]
	BatchDeleter[T]
	Transactioner
}

// HTTPHandler defines an interface for types that can register HTTP handlers.
type ModelHandlerer[T any] interface {
	Modeller[T]
	NewModelHandlerer(Modeller[T]) ModelHandlerer[T]
	RegisterHandlers(mux *http.ServeMux)
	Handle_Get_One(w http.ResponseWriter, r *http.Request)
	Handle_Get_List(w http.ResponseWriter, r *http.Request)
	Handle_Post(w http.ResponseWriter, r *http.Request)
	Handle_Put(w http.ResponseWriter, r *http.Request)
	Handle_Delete(w http.ResponseWriter, r *http.Request)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type UserModelHandler struct {
	ModelHandlerer[User]
	db *gorm.DB
}

func (h *UserModelHandler) NewModelHandlerer(modeller Modeller[User]) ModelHandlerer[User] {
	user_db, err := gorm.Open(sqlite.Open("database/user.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return &UserModelHandler{
		ModelHandlerer[User]{modeller}, 
		user_db,
	}
}

func (h *UserModelHandler) Create(u *User) error {
	UserDB := ConnectToUserDB()
	tx := UserDB.
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *UserModelHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/create", h.Handle_Post)
	mux.HandleFunc("/read", h.Handle_Get_One)
	mux.HandleFunc("/list", h.Handle_Get_List)
	mux.HandleFunc("/update", h.Handle_Put)
	mux.HandleFunc("/delete", h.Handle_Delete)
}

func (h *UserModelHandler) Handle_Get_One(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Handle the request, possibly using h.Collection.Read(...)
}

func (h *UserModelHandler) Handle_Get_List(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Handle the request, possibly using h.Collection.List(...)
}

func (h *UserModelHandler) Handle_Post(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Handle the request, possibly using h.Collection.Create(...)
}

func (h *UserModelHandler) Handle_Put(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Handle the request, possibly using h.Collection.Update(...)
}

func (h *UserModelHandler) Handle_Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	// Handle the request, possibly using h.Collection.Delete(...)
}

func (h *UserModelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle the request, possibly using h.Collection.Create(...)
}
