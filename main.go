package main

import (
	"fmt"
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

	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":8080", nil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
