package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// Template data structure
type PageData struct {
	Title   string
	Content string
}

func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Using template for home page
	data := PageData{
		Title:   "Welcome to My Website",
		Content: "This is the home page created with Go templates!",
	}

	tmpl, err := template.ParseFiles(filepath.Join("components", "home.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	// Serve a static HTML file for the about page
	http.ServeFile(w, r, filepath.Join("static", "about.html"))
}

func main() {

	// Create a file server to serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	// Handle requests to /static/ by stripping the prefix and serving from fs
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register route handlers
	http.HandleFunc("/", homePage)
	http.HandleFunc("/about", aboutPage)

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}

}
