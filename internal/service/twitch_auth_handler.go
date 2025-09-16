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
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
