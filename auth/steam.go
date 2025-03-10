package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

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

// OpenID nonceStore and discovery
var nonceStore = openid.NewSimpleNonceStore()
var discoveryCache = openid.NewSimpleDiscoveryCache()

// HandleSteamLogin processes Steam OpenID authentication
func HandleSteamLogin(w http.ResponseWriter, r *http.Request) {
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
		steamUser, err := GetSteamUserInfo(steamID)
		if err != nil {
			http.Error(w, "Failed to get Steam user info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create a session for the user
		SetSessionCookie(w, steamUser)

		// Redirect to home page after successful login
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// If this is the initial login request, redirect to Steam
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

// GetSteamUserInfo fetches user information from the Steam API
func GetSteamUserInfo(steamID string) (*SteamUser, error) {
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

// SetSessionCookie creates a simple cookie to remember the user
func SetSessionCookie(w http.ResponseWriter, user *SteamUser) {
	cookie := &http.Cookie{
		Name:     "steam_user",
		Value:    url.QueryEscape(user.SteamID + "|" + user.PersonaName),
		Path:     "/",
		MaxAge:   3600 * 24 * 7, // 1 week
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// ClearSession removes the user's session cookie
func ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "steam_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// GetCurrentUser gets the current user from the cookie
func GetCurrentUser(r *http.Request) *SteamUser {
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
	user, err := GetSteamUserInfo(steamID)
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
