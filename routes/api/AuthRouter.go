package api

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
)

type AuthRouter struct {
	mux     *mux.Router
	Context context.Context
}

func NewAuthRouter(context context.Context) *AuthRouter {
	ar := &AuthRouter{
		mux:     mux.NewRouter(),
		Context: context,
	}
	ar.routes()

	fmt.Println("NewAuthRouter")
	return ar
}

func (ar *AuthRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AuthRouter.ServeHTTP" + r.URL.Path)
	ar.mux.ServeHTTP(w, r)
}

func (ar *AuthRouter) routes() {
	ar.mux.HandleFunc("/api/auth/login", ar.HandleLogin).Methods("POST")
	ar.mux.HandleFunc("/logout", ar.HandleLogout).Methods("POST")
}

func (ar *AuthRouter) HandleLogin(w http.ResponseWriter, r *http.Request) {
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
	session.Save(r, w)

	// Update user's last login time
	user.LastLoginAt = time.Now()
	userDB.UpdateUser(user)

	http.Redirect(w, r, "/dashboard/", http.StatusSeeOther)
}

func (ar *AuthRouter) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
