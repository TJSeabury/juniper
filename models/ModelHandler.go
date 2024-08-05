package models

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ModelHandler[T any] struct {
	db         *gorm.DB
	Mux        *http.ServeMux
	TypeName   string
	model      *T
	jsonMapper func(map[string]interface{}) (T, error)
}

func NewModelHandler[T any](
	model *T,
	jsonMapper func(map[string]interface{}) (T, error),
	databaseLocation string,
	databaseConnectionConfig *gorm.Config,
	router *http.ServeMux,
	context context.Context,
	allowedOrigins []string,
	allowedMethods []string,
) *ModelHandler[T] {
	name := reflect.TypeOf(*model).Name()
	name = strings.ToLower(name) + "s"

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

	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   allowedOrigins,
	// 	AllowCredentials: true,
	// 	AllowedMethods:   allowedMethods,
	// })

	// c.Handler(modelHandler.r)

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
	handler.Mux.ServeHTTP(w, r)
}

func (handler *ModelHandler[T]) RegisterHandlers(context context.Context) {
	handler.Mux.HandleFunc(
		"GET /api/"+handler.TypeName+"/{slug}",
		handler.Handle_Get_One(handler.TypeName),
	)
	handler.Mux.HandleFunc(
		"GET /api/"+handler.TypeName+"/",
		handler.Handle_Get_List,
	)
	handler.Mux.HandleFunc(
		"POST /api/"+handler.TypeName+"/",
		handler.Handle_Post(handler.jsonMapper),
	)
	handler.Mux.HandleFunc(
		"PUT /api/"+handler.TypeName+"/{slug}",
		handler.Handle_Put(
			"id",
			handler.jsonMapper,
		),
	)
	handler.Mux.HandleFunc(
		"DELETE /api/"+handler.TypeName+"/{slug}",
		handler.Handle_Delete(
			"id",
		),
	)

	notFoundPatterns := []string{
		"PUT /api/" + handler.TypeName + "/",
		"DELETE /api/" + handler.TypeName + "/",
		"GET /api/" + handler.TypeName + "/{slug}/...",
		"POST /api/" + handler.TypeName + "/{slug}/...",
		"PUT /api/" + handler.TypeName + "/{slug}/...",
		"DELETE /api/" + handler.TypeName + "/{slug}/...",
	}
	for _, pattern := range notFoundPatterns {
		handler.Mux.HandleFunc(pattern, handler.Handle_NotFound)
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
	jsonMapper func(map[string]interface{}) (T, error),
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
		model, err := jsonMapper(data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
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
	jsonMapper func(map[string]interface{}) (T, error),
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

		modelData, err := jsonMapper(data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

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
