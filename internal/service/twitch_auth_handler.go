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
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
