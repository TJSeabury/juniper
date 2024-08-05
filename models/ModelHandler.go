package models

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"reflect"

	"github.com/rs/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ModelHandler[T any] struct {
	db         *gorm.DB
	r          *http.ServeMux
	typeName   string
	model      *T
	jsonMapper func(map[string]interface{}) T
}

func NewModelHandler[T any](
	model *T,
	jsonMapper func(map[string]interface{}) T,
	databaseLocation string,
	databaseConnectionConfig *gorm.Config,
	router *http.ServeMux,
	context context.Context,
	allowedOrigins []string,
	allowedMethods []string,
) *ModelHandler[T] {
	name := reflect.TypeOf(*model).Name()
	log.Println("name", name)
	post_db, err := gorm.Open(
		sqlite.Open(databaseLocation),
		databaseConnectionConfig,
	)
	if err != nil {
		panic("failed to connect database")
	}
	post_db.AutoMigrate(model)

	// Enable WAL mode for SQLite
	err = post_db.Exec("PRAGMA journal_mode=WAL;").Error
	if err != nil {
		panic("failed to enable WAL mode")
	}

	modelHandler := &ModelHandler[T]{
		post_db,
		router,
		name,
		model,
		jsonMapper,
	}

	modelHandler.RegisterHandlers(context)

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedMethods:   allowedMethods,
	})

	c.Handler(modelHandler.r)

	return modelHandler
}

func (handler *ModelHandler[T]) Create(model *T) error {
	tx := handler.db.Create(&model)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) Read(model *T) error {
	tx := handler.db.First(&model)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) Update(model *T) error {
	tx := handler.db.Save(&model)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) Delete(model *T) error {
	tx := handler.db.Delete(&model)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) Find(model *T) error {
	tx := handler.db.Find(model)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) List() ([]T, error) {
	var models []T
	handler.db.Find(&models)
	return models, nil
}

func (handler *ModelHandler[T]) Count() (int64, error) {
	var count int64
	handler.db.Count(&count)
	return count, nil
}

func (handler *ModelHandler[T]) Exists(model *T) (bool, error) {
	var exists bool
	tx := handler.db.Where(model).First(&exists)
	if tx.Error != nil {
		return exists, tx.Error
	}
	return exists, nil
}

func (handler *ModelHandler[T]) BatchCreate(items []T) error {
	tx := handler.db.Create(&items)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) BatchUpdate(items []T) error {
	tx := handler.db.Save(&items)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) BatchDelete(ids []int) error {
	tx := handler.db.Delete(&ids)
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (handler *ModelHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.r.ServeHTTP(w, r)
}

func (handler *ModelHandler[T]) RegisterHandlers(context context.Context) {
	handler.r.HandleFunc(
		"GET /api/"+handler.typeName+"/{slug}",
		handler.Handle_Get_One(handler.typeName),
	)
	handler.r.HandleFunc(
		"GET /api/"+handler.typeName+"/",
		handler.Handle_Get_List,
	)
	handler.r.HandleFunc(
		"POST /api/"+handler.typeName+"/",
		handler.Handle_Post(handler.jsonMapper),
	)
	handler.r.HandleFunc(
		"PUT /api/"+handler.typeName+"/{slug}",
		handler.Handle_Put(
			"id",
			handler.jsonMapper,
		),
	)
	handler.r.HandleFunc(
		"DELETE /api/"+handler.typeName+"/{slug}",
		handler.Handle_Delete(
			"id",
		),
	)

	notFoundPatterns := []string{
		"PUT /api/" + handler.typeName + "/",
		"DELETE /api/" + handler.typeName + "/",
		"GET /api/" + handler.typeName + "/{slug}/...",
		"POST /api/" + handler.typeName + "/{slug}/...",
		"PUT /api/" + handler.typeName + "/{slug}/...",
		"DELETE /api/" + handler.typeName + "/{slug}/...",
	}
	for _, pattern := range notFoundPatterns {
		handler.r.HandleFunc(pattern, handler.Handle_NotFound)
	}

}

func (handler *ModelHandler[T]) Handle_Get_One(
	pathParamName string,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParamValue := r.PathValue(pathParamName)
		model := handler.model
		tx := handler.db.First(&model, pathParamValue)
		if tx.Error != nil {
			http.Error(w, tx.Error.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model)
	}
}

func (handler *ModelHandler[T]) Handle_NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Endpoint not found."})
}

func (handler *ModelHandler[T]) Handle_Get_List(
	w http.ResponseWriter,
	r *http.Request,
) {
	models := make([]T, 0)
	handler.db.Find(&models)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models)
}

func (handler *ModelHandler[T]) Handle_Post(
	jsonMapper func(map[string]interface{}) T,
) func(w http.ResponseWriter, r *http.Request) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		model := jsonMapper(data)
		err = handler.Create(&model)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model)
	}
}

func (handler *ModelHandler[T]) Handle_Put(
	pathParamName string,
	jsonMapper func(map[string]interface{}) T,
) func(w http.ResponseWriter, r *http.Request) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		pathParamValue := r.PathValue(pathParamName)
		model := handler.model
		tx := handler.db.First(&model, pathParamValue)
		if tx.Error != nil {
			http.Error(w, tx.Error.Error(), http.StatusNotFound)
			return
		}

		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		modelData := jsonMapper(data)

		err = handler.Update(&modelData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model)
	}
}

func (handler *ModelHandler[T]) Handle_Delete(
	pathParamName string,
) func(w http.ResponseWriter, r *http.Request) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		pathParamValue := r.PathValue(pathParamName)
		model := handler.model
		tx := handler.db.First(model, pathParamValue)
		if tx.Error != nil {
			http.Error(w, tx.Error.Error(), http.StatusNotFound)
			return
		}

		err := handler.Delete(model)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model)
	}
}
