package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/yohcop/openid-go"
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

// OpenID nonceStore and discovery
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

func loginPage(w http.ResponseWriter, r *http.Request) {
	// If this is a callback from Steam
	if r.URL.Query().Get("openid.mode") != "" {
		// Verify the Steam authentication
		id, err := openid.Verify(
			"http://"+r.Host+r.URL.String(),
			discoveryCache,
			nonceStore)

		if err != nil {
			http.Error(w, "Steam authentication failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Extract the Steam ID from the OpenID response
		steamID := strings.TrimPrefix(id, "https://steamcommunity.com/openid/id/")

		// Get user info from Steam API
		steamUser, err := getSteamUserInfo(steamID)
		if err != nil {
			http.Error(w, "Failed to get Steam user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a session for the user (simplified for this example)
		// In a real app, you'd use a session management library
		setSessionCookie(w, steamUser)

		// Redirect to home page after successful login
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// If this is the initial login request, redirect to Steam
	// Instead of using the openid.RedirectURL function, construct the URL manually
	realm := "http://" + r.Host
	returnTo := realm + "/login"

	// Construct the Steam OpenID URL directly
	authURL := "https://steamcommunity.com/openid/login" +
		"?openid.ns=http://specs.openid.net/auth/2.0" +
		"&openid.mode=checkid_setup" +
		"&openid.return_to=" + url.QueryEscape(returnTo) +
		"&openid.realm=" + url.QueryEscape(realm) +
		"&openid.identity=http://specs.openid.net/auth/2.0/identifier_select" +
		"&openid.claimed_id=http://specs.openid.net/auth/2.0/identifier_select"

	// Redirect the user to Steam for authentication
	http.Redirect(w, r, authURL, http.StatusFound)
}

// getSteamUserInfo fetches user information from the Steam API
func getSteamUserInfo(steamID string) (*SteamUser, error) {
	steamAPIKey := os.Getenv("STEAM_API_KEY")
	if steamAPIKey == "" {
		return nil, fmt.Errorf("STEAM_API_KEY environment variable not set")
	}

	url := fmt.Sprintf(
		"https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		steamAPIKey, steamID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var steamResp SteamResponse
	err = json.Unmarshal(body, &steamResp)
	if err != nil {
		return nil, err
	}

	if len(steamResp.Response.Players) == 0 {
		return nil, fmt.Errorf("no player data returned from Steam")
	}

	return &steamResp.Response.Players[0], nil
}

// setSessionCookie creates a simple cookie to remember the user
// In a real application, you would use a proper session management library
func setSessionCookie(w http.ResponseWriter, user *SteamUser) {
	// Create a simple cookie with the user's Steam ID and name
	// This is NOT secure for production - use a proper session library instead
	cookie := &http.Cookie{
		Name:     "steam_user",
		Value:    url.QueryEscape(user.SteamID + "|" + user.PersonaName),
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 1 week
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	cookie := &http.Cookie{
		Name:     "steam_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	// Redirect to home page after logout
	http.Redirect(w, r, "/", http.StatusFound)
}

// getCurrentUser gets the current user from the cookie
// In a real application, you would use a proper session management library
func getCurrentUser(r *http.Request) *SteamUser {
	cookie, err := r.Cookie("steam_user")
	if err != nil {
		return nil
	}

	value, _ := url.QueryUnescape(cookie.Value)
	parts := strings.Split(value, "|")
	if len(parts) != 2 {
		return nil
	}

	steamID := parts[0]

	// Fetch the complete user profile from Steam API
	user, err := getSteamUserInfo(steamID)
	if err != nil {
		log.Printf("Error fetching Steam user info: %v", err)
		// Fall back to basic info from cookie if API call fails
		return &SteamUser{
			SteamID:     steamID,
			PersonaName: parts[1],
		}
	}

	return user
}

// Update the template cache to include the login page
var templates = template.Must(template.ParseFiles(
	filepath.Join("components", "header.html"),
	filepath.Join("components", "footer.html"),
	filepath.Join("components", "home.html"),
	filepath.Join("components", "login.html"),
	filepath.Join("components", "search.html"),
))

// Update the homePage function to display user info if logged in
// Update the homePage function to avoid multiple WriteHeader calls
func homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get the current user if logged in
	user := getCurrentUser(r)

	// Create data for the template
	data := struct {
		User *SteamUser
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

// API endpoint to get current time
func getCurrentTime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"time": "%s"}`, time.Now().Format("2006-01-02 15:04:05"))
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// Get the current user if logged in
	user := getCurrentUser(r)

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
		User: user,
	}

	err := templates.ExecuteTemplate(w, "search.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		return
	}
}

func main() {
	// Create a file server to serve static files from the "static" directory
	fs := http.FileServer(http.Dir("static"))
	// Handle requests to /static/ by stripping the prefix and serving from fs
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register route handlers
	http.HandleFunc("/", homePage)
	http.HandleFunc("/api/time", getCurrentTime)
	http.HandleFunc("/login", loginPage)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/search", searchHandler)

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
