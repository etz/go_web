package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/etz/go_web/auth"
	"github.com/joho/godotenv"
)

// SteamUser represents a Steam user profile
type SteamUser struct {
	SteamID      string `json:"steamid"`
	PersonaName  string `json:"personaname"`
	ProfileURL   string `json:"profileurl"`
	Avatar       string `json:"avatar"`
	AvatarMedium string `json:"avatarmedium"`
	AvatarFull   string `json:"avatarfull"`
}

// SteamResponse represents the response from Steam API
type SteamResponse struct {
	Response struct {
		Players []SteamUser `json:"players"`
	} `json:"response"`
}

// Load environment variables from .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// Update the template cache to include the login page
var templates = template.Must(template.ParseFiles(
	filepath.Join("components", "header.html"),
	filepath.Join("components", "footer.html"),
	filepath.Join("components", "home.html"),
	filepath.Join("components", "login.html"),
	filepath.Join("components", "search.html"),
	filepath.Join("components", "terms.html"),
	filepath.Join("components", "privacy.html"),
))

// Update the homePage function to display user info if logged in
// Update the homePage function to avoid multiple WriteHeader calls
func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get the current user if logged in
	user := auth.GetCurrentUser(r)

	// Create data for the template
	data := struct {
		User *auth.SteamUser
	}{
		User: user,
	}

	// Use the template instead of serving a static file
	err := templates.ExecuteTemplate(w, "home.html", data)
	if err != nil {
		// Don't call http.Error if you've already written to the response
		log.Printf("Template execution error: %v", err)
		// Instead of http.Error, just log the error
		return
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// Get the current user if logged in
	user := auth.GetCurrentUser(r)

	// In a real application, you would search your database or content here
	// For now, we'll just return the search query
	data := struct {
		Query   string
		Results []string
		User    *SteamUser
	}{
		Query: query,
		Results: []string{
			"Result 1 for: " + query,
			"Result 2 for: " + query,
			"Result 3 for: " + query,
		},
		User: (*SteamUser)(user),
	}

	err := templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}
}

// Handler for logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.ClearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Example: API endpoint to get current time
func getCurrentTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"time": "%s"}`, time.Now().Format("2006-01-02 15:04:05"))
}

// Add these handler functions

func termsHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)

	data := struct {
		User *auth.SteamUser
	}{
		User: user,
	}

	err := templates.ExecuteTemplate(w, "terms.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}
}

func privacyHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.GetCurrentUser(r)

	data := struct {
		User *auth.SteamUser
	}{
		User: user,
	}

	err := templates.ExecuteTemplate(w, "privacy.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}
}

// In the main function, add these routes
func main() {
	// Create a file server to serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	// Handle requests to /static/ by stripping the prefix and serving from fs
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register route handlers
	http.HandleFunc("/", homePage)
	http.HandleFunc("/api/time", getCurrentTime)
	http.HandleFunc("/login", auth.HandleSteamLogin)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/terms", termsHandler)
	http.HandleFunc("/privacy", privacyHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
