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
)

type TwitchUser struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type TwitchAuthHandler struct {
	clientID     string
	clientSecret string
	redirectURL  string
}

func NewTwitchAuthHandler() *TwitchAuthHandler {
	return &TwitchAuthHandler{
		clientID:     os.Getenv("TWITCH_CLIENT_ID"),
		clientSecret: os.Getenv("TWITCH_CLIENT_SECRET"),
		redirectURL:  os.Getenv("TWITCH_REDIRECT_URL"),
	}
}

func (t *TwitchAuthHandler) GenerateAuthURL() string {
	state := generateState()
	params := url.Values{
		"client_id":     {t.clientID},
		"redirect_uri":  {t.redirectURL},
		"response_type": {"code"},
		"scope":         {"user:read:email"},
		"state":         {state},
	}
	return "https://id.twitch.tv/oauth2/authorize?" + params.Encode()
}

func (t *TwitchAuthHandler) ExchangeCode(ctx context.Context, code string) (string, error) {
	data := url.Values{
		"client_id":     {t.clientID},
		"client_secret": {t.clientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {t.redirectURL},
	}

	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AcessToken string `json:"acess_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AcessToken, nil
}

func (t *TwitchAuthHandler) GetUser(ctx context.Context, accessToken string) (*TwitchUser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitch.tv/helix/users", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer"+accessToken)
	req.Header.Set("Client-Id", t.clientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data []TwitchUser `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("No user data")
	}
	return &result.Data[0], nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
