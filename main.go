package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

// Template data structure
type PageData struct {
	Title       string
	Content     string
	CurrentTime string
	Items       []string
	User        *User
}

// User represents a user of the website
type User struct {
	Name  string
	Email string
	Role  string
}

// Create a template cache
// Update the template cache to include the contact page
var templates = template.Must(template.ParseFiles(
	filepath.Join("components", "header.html"),
	filepath.Join("components", "footer.html"),
	filepath.Join("components", "home.html"),
	filepath.Join("components", "contact.html"),
))

// Add a contact page handler
func contactPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Contact Us",
		Content: "Get in touch with our team",
	}

	err := templates.ExecuteTemplate(w, "contact.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Using template for home page with dynamic data
	data := PageData{
		Title:       "Welcome to My Website",
		Content:     "This is the home page created with Go templates!",
		CurrentTime: time.Now().Format("2006-01-02 15:04:05"),
		Items:       []string{"Item 1", "Item 2", "Item 3", "Item 4"},
		User: &User{
			Name:  "John Doe",
			Email: "john@example.com",
			Role:  "Admin",
		},
	}

	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	// Serve a static HTML file for the about page
	http.ServeFile(w, r, filepath.Join("static", "about.html"))
}

// API endpoint to get current time
func getCurrentTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"time": "%s"}`, time.Now().Format("2006-01-02 15:04:05"))
}

func main() {
	// Create a file server to serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	// Handle requests to /static/ by stripping the prefix and serving from fs
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register route handlers
	http.HandleFunc("/", homePage)
	http.HandleFunc("/about", aboutPage)
	http.HandleFunc("/api/time", getCurrentTime)
	http.HandleFunc("/contact", contactPage)

	// Start the server
	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
