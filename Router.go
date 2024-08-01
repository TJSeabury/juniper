package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/auth"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/views/dashboard"
	"pioneerwebworks.com/juniper/views/partials"
	"pioneerwebworks.com/juniper/views/public"
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

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Router.ServeHTTP" + req.URL.Path)
	router.Mux.ServeHTTP(w, req)
}

func (router *Router) routes() {
	// API routes
	router.Mux.HandleFunc("/api/auth/login", router.api_auth_login).Methods("POST")
	router.Mux.HandleFunc("/api/auth/logout", router.api_auth_logout).Methods("POST")
	router.Mux.HandleFunc("/api/auth/status", router.api_auth_status).Methods("GET")
	router.Mux.HandleFunc("/api/auth/register", router.api_auth_register).Methods("POST")
	router.Mux.HandleFunc("/api/auth/verify-email", router.api_auth_verify_email).Methods("GET")
	/**
	 * @todo
	 * - auth forgot password
	 * - auth reset password
	 * - auth change password
	 * - auth change email
	 * - auth change username
	 * - auth delete account
	 * - auth update account
	 * - auth get account
	 */

	postModelHandler := models.NewPostModelHandler(
		router.Mux,
		router.Context,
	)

	router.Mux.PathPrefix("/api/post").Handler(
		//auth.WithAuth(postModelHandler),
		postModelHandler,
	)

	router.Mux.PathPrefix("/dashboard").Handler(
		auth.WithAuth(&DashboardHandler{Context: router.Context}),
	)
	router.Mux.PathPrefix("/").Handler(
		&PublicHandler{Context: router.Context},
	)
}

func (router *Router) api_auth_verify_email(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the token from the URL query parameter
	username := r.URL.Query().Get("username")
	token := r.URL.Query().Get("token")

	// Authenticate user
	userDB := models.ConnectToUserDB()

	user := userDB.FindByUsername(username)

	tokenIsValid := user.CheckEmailToken(token)

	if !tokenIsValid {
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
	user.EmailVerified = true
	userDB.UpdateUser(user)

	// return success
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\": \"Success\"}"))
}

func (router *Router) api_auth_register(w http.ResponseWriter, r *http.Request) {
	userDB := models.ConnectToUserDB()

	type registerForm struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		Email     string `json:"email"`
		Forename  string `json:"forename"`
		Surname   string `json:"surname"`
		Phone     string `json:"phone"`
		Birthdate string `json:"birthdate"`
	}

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON data into the struct
	var data registerForm
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parsing JSON body", http.StatusBadRequest)
		log.Println(err)
		return
	}

	// Use the data
	log.Printf("Received: %+v", data)

	// Check if user already exists
	var user models.User
	userDB.DB.First(&user, "username = ?", data.Username)
	if user.Username != "" {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	hashedPassword, err := models.HashPassword(data.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(data.Email)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	hashedEmailToken, err := models.HashEmailToken(data.Email)
	if err != nil {
		http.Error(w, "Error hashing email token", http.StatusInternalServerError)
		return
	}

	parsedBirthdate, err := time.Parse("2006-01-02", data.Birthdate)
	if err != nil {
		http.Error(w, "Error parsing birthdate", http.StatusInternalServerError)
		return
	}

	user = models.User{
		Username:      data.Username,
		Password:      hashedPassword,
		Email:         data.Email,
		EmailToken:    hashedEmailToken,
		Forename:      data.Forename,
		Surname:       data.Surname,
		PhoneNumber:   data.Phone,
		Birthdate:     parsedBirthdate,
		EmailVerified: false,
		PhoneVerified: false,
		UserRole:      "user",
	}
	user.ID, err = userDB.CreateUser(&user)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// URL encode the token
	token = url.QueryEscape(token)

	// Send verification email
	email := Email{
		To:      []string{user.Email},
		From:    APP_CONFIG["SMTP_USERNAME"],
		Subject: "Verify your email address",
		Body:    "Please verify your email address by clicking the link below:\n\n" + APP_CONFIG["SITE_URL"] + "/verify?token=" + token + "&username=" + user.Username,
	}

	err = GlobalMailer.Send(email)
	if err != nil {
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	session, _ := auth.Store.Get(r, "juniper-session")

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
	w.Write([]byte("{\"message\": \"Success\"}"))
}

func (router *Router) api_auth_login(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte("{\"message\": \"Success\"}"))
}

func (router *Router) api_auth_logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\": \"Success\"}"))
}

func (router *Router) api_auth_status(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("{\"message\": \"Unauthorized\"}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"message\": \"Authenticated\"}"))
}

type DashboardHandler struct {
	Context context.Context
}

func (dh *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	posts := []models.Post{}
	post_db.Find(&posts)
	public.App(
		dashboard.Dashboard(posts),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(dh.Context, w)
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

type PublicHandler struct {
	Context context.Context
}

func (ph *PublicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		ph.public_Home(w, r)
	case "/about":
		ph.public_About(w, r)
	case "/verify":
		ph.public_Verify(w, r)
	case "/register":
		ph.public_Register(w, r)
	case "/login":
		ph.public_Login(w, r)
	default:
		ph.public_Blog(w, r)
	}
}

func (ph *PublicHandler) public_Home(w http.ResponseWriter, r *http.Request) {
	c := public.Paragraph("Home page content.")
	public.App(
		c,
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_About(w http.ResponseWriter, r *http.Request) {
	public.App(
		public.Page_About(),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Verify(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")
	userDB := models.ConnectToUserDB()

	// Get the token from the URL query parameter
	username := r.URL.Query().Get("username")
	token := r.URL.Query().Get("token")

	// URL decode the token
	token, _ = url.QueryUnescape(token)

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Println("auth", auth)
	log.Println("ok", ok)

	user := userDB.FindByUsername(username)

	// Check that username matches the session user
	// sessionUserID, _ := session.Values["userID"].(int)
	// log.Println("sessionUserID", sessionUserID)
	// log.Println("user.ID", user.ID)
	// if user.ID != uint(sessionUserID) {
	// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	// 	return
	// }

	tokenIsValid := user.CheckEmailToken(token)

	log.Println("tokenIsValid", tokenIsValid)

	// Update user's last login time
	user.LastLoginAt = time.Now()
	user.EmailVerified = true
	userDB.UpdateUser(user)

	public.App(
		partials.Verify(tokenIsValid),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Register(w http.ResponseWriter, r *http.Request) {
	public.App(
		partials.Register(),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Login(w http.ResponseWriter, r *http.Request) {
	public.App(
		partials.Login(),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Blog(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	//Get the post ID from the URL path
	postID := r.URL.Path[len("/blog/"):]
	post := models.Post{}
	post_db.First(&post, postID)
	public.App(
		public.Post(post),
		public.Header(),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}
