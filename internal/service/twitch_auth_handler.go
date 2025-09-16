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
