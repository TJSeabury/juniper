package main

import (
	"log"
	"net/http"
	"os"

	"pioneerwebworks.com/juniper/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	user_db, err := gorm.Open(sqlite.Open("database/user.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	user_db.AutoMigrate(&models.User{})

	post_db, err := gorm.Open(sqlite.Open("database/post.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	post_db.AutoMigrate(&models.Post{})

	// Serve static files from public/media under the /media URL path
	mediaFs := http.FileServer(http.Dir("public/media"))
	http.Handle("/media/", http.StripPrefix("/media/", mediaFs))
	stylesFs := http.FileServer(http.Dir("public/styles"))
	http.Handle("/styles/", http.StripPrefix("/styles/", stylesFs))
	scriptsFs := http.FileServer(http.Dir("public/scripts"))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", scriptsFs))
	fontsFs := http.FileServer(http.Dir("public/fonts"))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", fontsFs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
