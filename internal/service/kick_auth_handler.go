package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type KickUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type PKCEData struct {
	CodeVerifier  string
	CodeChallenge string
	State         string
}

type KickAuthHandler struct {
	clientID     string
	clientSecret string
	redirectURL  string
	stateStore   *StateStore
	pkceStore    map[string]string // state -> codeVerifier mapping
	pkceMutex    sync.RWMutex
}

func NewKickAuthHandler() *KickAuthHandler {
	return &KickAuthHandler{
		clientID:     os.Getenv("KICK_CLIENT_ID"),
		clientSecret: os.Getenv("KICK_CLIENT_SECRET"),
		redirectURL:  os.Getenv("KICK_REDIRECT_URL"),
		stateStore:   NewStateStore(),
		pkceStore:    make(map[string]string),
	}
}

func (k *KickAuthHandler) GenerateAuthURL() string {
	state := generateState()
	codeVerifier := generateCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Store state and code verifier
	k.stateStore.Add(state)
	k.pkceMutex.Lock()
	k.pkceStore[state] = codeVerifier
	k.pkceMutex.Unlock()

	params := url.Values{
		"client_id":             {k.clientID},
		"redirect_uri":          {k.redirectURL},
		"response_type":         {"code"},
		"scope":                 {"user:read"},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
	}

	return "https://id.kick.com/oauth/authorize?" + params.Encode()
}

func (k *KickAuthHandler) ValidateState(state string) bool {
	return k.stateStore.Validate(state)
}

func (k *KickAuthHandler) ExchangeCode(ctx context.Context, code string, state string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("authorization code is required")
	}

	var codeVerifier string
	if state == "" {
		return "", fmt.Errorf("state parameter is required for PKCE exchange")
	}

	k.pkceMutex.Lock()
	codeVerifier, exists := k.pkceStore[state]
	if !exists {
		k.pkceMutex.Unlock()
		return "", fmt.Errorf("invalid state or PKCE data not found")
	}
	// Clean up PKCE data
	delete(k.pkceStore, state)
	k.pkceMutex.Unlock()

	data := url.Values{
		"client_id":     {k.clientID},
		"client_secret": {k.clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {k.redirectURL},
		"code_verifier": {codeVerifier},
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://id.kick.com/oauth/token", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = data.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code with Kick: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Kick token exchange failed with status: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Kick response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("Kick API error: %s - %s", result.Error, result.ErrorDesc)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("no access token received from Kick")
	}

	return result.AccessToken, nil
}

func (k *KickAuthHandler) GetUser(ctx context.Context, accessToken string) (*KickUser, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.kick.com/public/v1/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from Kick: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Kick user API failed with status: %d", resp.StatusCode)
	}

	// Kick API returns users in a data array format
	var response struct {
		Data []struct {
			Email          string `json:"email"`
			Name           string `json:"name"`
			ProfilePicture string `json:"profile_picture"`
			UserID         int64  `json:"user_id"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode Kick user response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no user data received from Kick")
	}

	// Get the first user (should be the authenticated user when no ID is specified)
	kickUserData := response.Data[0]

	user := &KickUser{
		ID:          fmt.Sprintf("%d", kickUserData.UserID),
		Username:    kickUserData.Name,
		DisplayName: kickUserData.Name,
		Email:       kickUserData.Email,
	}

	if user.ID == "" {
		return nil, fmt.Errorf("no user ID received from Kick")
	}

	return user, nil
}

func (k *KickAuthHandler) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	if refreshToken == "" {
		return "", fmt.Errorf("refresh token is required")
	}

	data := url.Values{
		"client_id":     {k.clientID},
		"client_secret": {k.clientSecret},
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://id.kick.com/oauth/token", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = data.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token with Kick: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Kick token refresh failed with status: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Kick refresh response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("Kick refresh API error: %s - %s", result.Error, result.ErrorDesc)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("no access token received from Kick refresh")
	}

	return result.AccessToken, nil
}

// generateCodeVerifier creates a random string for PKCE code verifier
func generateCodeVerifier() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// generateCodeChallenge creates SHA256 hash of code verifier for PKCE code challenge
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}
