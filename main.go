package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"pioneerwebworks.com/juniper/auth"
	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/routes"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
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
			Username: "admin",
			Password: hashedPassword,
			UserRole: "administrator",
		}
		user_db.Create(&adminUser)
	}

	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post_db.AutoMigrate(&models.Post{})

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

	router := routes.NewRouter(
		context.Background(),
	)
	http.Handle("/", router)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
