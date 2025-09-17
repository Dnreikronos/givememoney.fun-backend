package utils

type StreamerProvider string
type WalletProvider string

const (
	ProviderTwitch         StreamerProvider = "twitch"
	ProviderKick           StreamerProvider = "kick"
	ProviderYoutube        StreamerProvider = "kick"
	MetamaskWalletProvider WalletProvider   = "metamask"
	PhantomWalletProvider  WalletProvider   = "phantom"
)

type ProviderUser struct {
	ID    string
	Name  string
	Email string
}
