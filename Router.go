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

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"pioneerwebworks.com/juniper/auth"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/views/dashboard"
	"pioneerwebworks.com/juniper/views/partials"
	"pioneerwebworks.com/juniper/views/public"
)

type Router struct {
	Mux             *http.ServeMux
	Context         context.Context
	APIRouter       http.Handler
	DashboardRouter http.Handler
	PublicRouter    http.Handler
}

func NewRouter(context context.Context) *Router {
	r := &Router{
		Mux:     http.NewServeMux(),
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
	router.Mux.HandleFunc("POST /api/auth/login", router.api_auth_login)
	router.Mux.HandleFunc("POST /api/auth/logout", router.api_auth_logout)
	router.Mux.HandleFunc("GET /api/auth/status", router.api_auth_status)
	router.Mux.HandleFunc("POST /api/auth/register", router.api_auth_register)
	router.Mux.HandleFunc("GET /api/auth/verify-email", router.api_auth_verify_email)
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

	router.Mux.Handle(
		"/dashboard",
		auth.WithAuth(&DashboardHandler{Context: router.Context}),
	)
	router.Mux.Handle(
		"/",
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

	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var data map[string]interface{}
	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parsing JSON body", http.StatusBadRequest)
		return
	}

	// Use the data
	log.Printf("Received: %+v", data)
	log.Println(data["username"])
	log.Println(data["password"])

	//csrf := data["csrf"].(string)
	//nonce := data["nonce"].(string)
	user := userDB.FindByUsername(data["username"].(string))
	passwordVerified := user.CheckPassword(data["password"].(string))

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

func getSessionUser(r *http.Request) models.User {
	session, _ := auth.Store.Get(r, "juniper-session")

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		return models.User{}
	}

	userID, _ := session.Values["userID"].(uint)

	userDB := models.ConnectToUserDB()
	user := models.User{}
	user, err := userDB.GetUser(userID)
	if err != nil {
		return models.User{}
	}

	return user
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
	posts, err := APP_DATA.PostHandler.List()
	if err != nil {
		panic(err)
	}
	user := getSessionUser(r)
	public.App(
		dashboard.Dashboard(
			APP_DATA,
			APP_DATA.ListHandlerfields(),
			posts,
		),
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(dh.Context, w)
}

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
	case "/logout":
		ph.public_Logout(w, r)
	case "/blog":
		ph.public_Blog(w, r)
	default:
		ph.public_404(w, r)
	}
}

func (ph *PublicHandler) public_Home(w http.ResponseWriter, r *http.Request) {
	c := public.Paragraph("Home page content.")
	user := getSessionUser(r)
	public.App(
		c,
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_About(w http.ResponseWriter, r *http.Request) {
	user := getSessionUser(r)
	public.App(
		public.Page_About(),
		public.Header(user),
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
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Register(w http.ResponseWriter, r *http.Request) {
	user := getSessionUser(r)
	public.App(
		partials.Register(),
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Login(w http.ResponseWriter, r *http.Request) {
	user := getSessionUser(r)
	public.App(
		partials.Login(uuid.NewString()),
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := auth.Store.Get(r, "juniper-session")
	session.Options = &sessions.Options{MaxAge: -1}
	session.Values["authenticated"] = false
	session.Values["userID"] = 0
	session.Save(r, w)

	public.App(
		public.Paragraph("You have been logged out."),
		public.Header(models.User{}),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_Blog(w http.ResponseWriter, r *http.Request) {
	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	user := getSessionUser(r)
	posts := []models.Post{}
	post_db.Find(&posts)
	public.App(
		public.Blog(posts),
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}

func (ph *PublicHandler) public_404(w http.ResponseWriter, r *http.Request) {
	user := getSessionUser(r)
	public.App(
		public.Page_404(),
		public.Header(user),
		public.Footer(),
		public.Head("Juniper"),
	).Render(ph.Context, w)
}
