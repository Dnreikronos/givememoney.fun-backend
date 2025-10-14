package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type TwitchUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type StateStore struct {
	states map[string]time.Time
	mutex  sync.RWMutex
}

func NewStateStore() *StateStore {
	store := &StateStore{
		states: make(map[string]time.Time),
	}
	// Cleanup expired states every 10 minutes
	go store.cleanup()
	return store
}

func (s *StateStore) Add(state string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.states[state] = time.Now().Add(10 * time.Minute)
}

func (s *StateStore) Validate(state string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	expiry, exists := s.states[state]
	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		delete(s.states, state)
		return false
	}

	delete(s.states, state)
	return true
}

func (s *StateStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for state, expiry := range s.states {
			if now.After(expiry) {
				delete(s.states, state)
			}
		}
		s.mutex.Unlock()
	}
}

type TwitchAuthHandler struct {
	clientID     string
	clientSecret string
	redirectURL  string
	stateStore   *StateStore
}

func NewTwitchAuthHandler() *TwitchAuthHandler {
	return &TwitchAuthHandler{
		clientID:     os.Getenv("TWITCH_CLIENT_ID"),
		clientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		redirectURL:  os.Getenv("TWITCH_REDIRECT_URL"),
		stateStore:   NewStateStore(),
	}
}

func (t *TwitchAuthHandler) GenerateAuthURL() string {
	state := generateState()
	// Store state for validation
	t.stateStore.Add(state)

	params := url.Values{
		"client_id":     {t.clientID},
		"redirect_uri":  {t.redirectURL},
		"response_type": {"code"},
		"scope":         {"user:read:email"},
		"state":         {state},
	}
	return "https://id.twitch.tv/oauth2/authorize?" + params.Encode()
}

func (t *TwitchAuthHandler) ValidateState(state string) bool {
	return t.stateStore.Validate(state)
}

func (t *TwitchAuthHandler) ExchangeCode(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("authorization code is required")
	}

	data := url.Values{
		"client_id":     {t.clientID},
		"client_secret": {t.clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {t.redirectURL},
	}

	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code with Twitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("twitch token exchange failed with status: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode Twitch response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("twitch API error: %s - %s", result.Error, result.ErrorDesc)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("no access token received from Twitch")
	}

	return result.AccessToken, nil
}

func (t *TwitchAuthHandler) GetUser(ctx context.Context, accessToken string) (*TwitchUser, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Client-Id", t.clientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user from Twitch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twitch user API failed with status: %d", resp.StatusCode)
	}

	var result struct {
		Data  []TwitchUser `json:"data"`
		Error struct {
			Error   string `json:"error"`
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Twitch user response: %w", err)
	}

	if result.Error.Error != "" {
		return nil, fmt.Errorf("Twitch user API error: %s", result.Error.Message)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no user data received from Twitch")
	}

	return &result.Data[0], nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
