package models

import (
	"net/http"
	"reflect"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type AppData struct {
	UserHandler *ModelHandler[User]
	PostHandler *ModelHandler[Post]
	// CommentHandler *models.ModelHandler[models.Comment]
	// CategoryHandler *models.ModelHandler[models.Category]
	// TagHandler *models.ModelHandler[models.Tag]
	// TaggingHandler *models.ModelHandler[models.Tagging]
	// SettingHandler *models.ModelHandler[models.Setting]
}

func (appData *AppData) ListHandlerfields() []string {
	fields := reflect.TypeOf(*appData)
	numFields := fields.NumField()
	var fieldNames []string
	for i := 0; i < numFields; i++ {
		field := fields.Field(i)
		fieldNames = append(fieldNames, field.Name)
	}
	return fieldNames
}

type Creater[T any] interface {
	Create(*T) error
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
	Find(T) ([]T, error)
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
	Lister[T]
	Counter[T]
	ExistChecker[T]
	BatchCreater[T]
	BatchUpdater[T]
	BatchDeleter[T]
}

type IModelHandler[T any] interface {
	Modeller[T]
	NewModelHandlerer(Modeller[T]) IModelHandler[T]
	RegisterHandlers(mux *http.ServeMux)
	Handle_Get_One(w http.ResponseWriter, r *http.Request)
	Handle_Get_List(w http.ResponseWriter, r *http.Request)
	Handle_Post(w http.ResponseWriter, r *http.Request)
	Handle_Put(w http.ResponseWriter, r *http.Request)
	Handle_Delete(w http.ResponseWriter, r *http.Request)
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type UserModelHandler struct {
	IModelHandler[User]
	db *gorm.DB
}

func (h *UserModelHandler) NewModelHandlerer(modeller IModelHandler[User]) *UserModelHandler {
	user_db, err := gorm.Open(sqlite.Open("database/user.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return &UserModelHandler{
		modeller,
		user_db,
	}
}

func (h *UserModelHandler) Create(u *User) error {
	tx := h.db.Create(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *UserModelHandler) Read(u *User) error {
	tx := h.db.First(&u, u.ID)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *UserModelHandler) Update(u *User) error {
	tx := h.db.Save(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *UserModelHandler) Delete(u *User) error {
	tx := h.db.Delete(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *UserModelHandler) Find(u *User) error {
	tx := h.db.Find(&u)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (h *UserModelHandler) List(u *User) ([]User, error) {
	var users []User
	tx := h.db.Find(&users, &u)
	if tx.Error != nil {
		return users, tx.Error
	}
	return users, nil
}

func (h *UserModelHandler) Count(u *User) (int64, error) {
	var count int64
	tx := h.db.Count(&count)
	if tx.Error != nil {
		return count, tx.Error
	}
	return count, nil
}

func (h *UserModelHandler) Exists(u *User) (bool, error) {
	var exists bool
	tx := h.db.Where("username = ?", u.Username).First(&exists)
	if tx.Error != nil {
		return exists, tx.Error
	}
	return exists, nil
}

func (h *UserModelHandler) BatchCreate(u []User) error {
	for _, user := range u {
		tx := h.db.Create(&user)
		if tx.Error != nil {
			return tx.Error
		}
	}
	return nil
}

func (h *UserModelHandler) BatchUpdate(u []User) error {
	for _, user := range u {
		tx := h.db.Save(&user)
		if tx.Error != nil {
			return tx.Error
		}
	}
	return nil
}

func (h *UserModelHandler) BatchDelete(u []int) error {
	var users []User
	tx := h.db.Find(&users)
	if tx.Error != nil {
		return tx.Error
	}
	for _, user := range users {
		tx := h.db.Delete(&user)
		if tx.Error != nil {
			return tx.Error
		}
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
