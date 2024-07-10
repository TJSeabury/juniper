package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"pioneerwebworks.com/juniper/models"
	"pioneerwebworks.com/juniper/views"

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c := views.Paragraph("Home page content.")
		views.App(
			c,
			views.Header(),
			views.Footer(),
			views.Head("Juniper"),
		).Render(context.Background(), w)
	})

	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		views.App(
			views.Page_About(),
			views.Header(),
			views.Footer(),
			views.Head("Juniper"),
		).Render(context.Background(), w)
	})

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
