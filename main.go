package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"pioneerwebworks.com/juniper/auth"
	"pioneerwebworks.com/juniper/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var APP_CONFIG map[string]string
var GlobalMailer Mailer

var APP_DATA models.AppData

func main() {
	envFile, _ := godotenv.Read(".env")

	// get the values from the environment variables from .env file
	APP_CONFIG = envFile
	smtpUsername := envFile["SMTP_USERNAME"]
	smtpPassword := envFile["SMTP_PASSWORD"]
	smtpHost := envFile["SMTP_HOST"]
	smtpPortRaw := envFile["SMTP_PORT"]
	smtpPort, err := strconv.Atoi(smtpPortRaw)
	if err != nil {
		log.Fatal(err)
	}

	GlobalMailer = Mailer{
		Host:     smtpHost,
		Port:     int(smtpPort),
		Username: smtpUsername,
		Password: smtpPassword,
	}
	GlobalMailer.Initialize(smtpUsername, smtpPassword, smtpHost)

	user_db, err := gorm.Open(sqlite.Open("database/user.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	user_db.AutoMigrate(&models.User{})

	// Initialize the user database with a default admin user
	var adminUser models.User
	user_db.First(&adminUser, "username = ?", "admin")
	if adminUser.Username == "" {
		hashedPassword, err := models.HashPassword("admin")
		if err != nil {
			panic("failed to hash password")
		}
		adminUser = models.User{
			Username:      "admin",
			Password:      hashedPassword,
			UserRole:      "administrator",
			EmailVerified: true,
		}
		user_db.Create(&adminUser)
	}

	// Initialize the session store
	auth.Init()

	// Serve static files from public/media under the /media URL path
	mediaFs := http.FileServer(http.Dir("public/media"))
	http.Handle("/media/", http.StripPrefix("/media/", mediaFs))
	stylesFs := http.FileServer(http.Dir("public/styles"))
	http.Handle("/styles/", http.StripPrefix("/styles/", stylesFs))
	scriptsFs := http.FileServer(http.Dir("public/scripts"))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", scriptsFs))
	fontsFs := http.FileServer(http.Dir("public/fonts"))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", fontsFs))

	router := NewRouter(
		context.Background(),
	)

	APP_DATA = models.AppData{
		UserHandler: models.NewModelHandler[models.User](
			&models.User{},
			models.UserJSONMapper,
			"database/user.db",
			&gorm.Config{},
			router.Mux,
			router.Context,
			[]string{APP_CONFIG["SITE_URL"]},
			[]string{"GET", "POST", "PUT", "DELETE"},
		),
		PostHandler: models.NewModelHandler[models.Post](
			&models.Post{},
			models.PostJSONMapper,
			"database/post.db",
			&gorm.Config{},
			router.Mux,
			router.Context,
			[]string{APP_CONFIG["SITE_URL"]},
			[]string{"GET", "POST", "PUT", "DELETE"},
		),
	}

	http.Handle("/", router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
