package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"pioneerwebworks.com/juniper/auth"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/routes/api"
)

type Router struct {
	Mux             *mux.Router
	Context         context.Context
	APIRouter       http.Handler
	DashboardRouter http.Handler
	PublicRouter    http.Handler
}

func NewRouter(context context.Context) *Router {
	r := &Router{
		Mux:     mux.NewRouter(),
		Context: context,
	}
	r.routes()
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Router.ServeHTTP" + req.URL.Path)
	r.mux.ServeHTTP(w, req)
}

func (r *Router) routes() {
	r.DashboardRouter = NewDashboardRouter(r.Context)
	r.PublicRouter = NewPublicRouter(r.Context)
	r.APIRouter = api.NewAPIRouter(r.Context)
	r.mux.HandleFunc("/api/auth/login", HandleLogin).Methods("POST")
	r.mux.PathPrefix("/api").Handler(r.APIRouter)
	r.mux.PathPrefix("/dashboard").Handler(auth.WithAuth(r.DashboardRouter))
	r.mux.PathPrefix("/").Handler(r.PublicRouter)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Authenticate user
	userDB := models.ConnectToUserDB()

	type loginForm struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON data into the struct
	var data loginForm
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parsing JSON body", http.StatusBadRequest)
		return
	}

	// Use the data
	log.Printf("Received: %+v", data)

	user := userDB.FindByUsername(data.Username)
	passwordVerified := user.CheckPassword(data.Password)

	if !passwordVerified {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set user as authenticated
	session.Values["userID"] = user.ID
	session.Values["authenticated"] = true
	log.Println(session.Values)
	session.Save(r, w)

	// Update user's last login time
	user.LastLoginAt = time.Now()
	userDB.UpdateUser(user)

	// return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success"))
}
