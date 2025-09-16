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
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
