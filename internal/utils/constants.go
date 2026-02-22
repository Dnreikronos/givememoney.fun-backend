package utils

type StreamerProvider string
type WalletProvider string
type Currency string

const (
	ProviderTwitch         StreamerProvider = "twitch"
	ProviderKick           StreamerProvider = "kick"
	ProviderYoutube        StreamerProvider = "youtube"
	ProviderEmail          StreamerProvider = "email"
	MetamaskWalletProvider WalletProvider   = "metamask"
	PhantomWalletProvider  WalletProvider   = "phantom"

	CurrencyETH  Currency = "ETH"
	CurrencySOL  Currency = "SOL"
	CurrencyUSDT Currency = "USDT"
	CurrencyUSDC Currency = "USDC"
)

// SupportedCurrencies lists currencies accepted for donations (native + stablecoins).
var SupportedCurrencies = []Currency{CurrencyETH, CurrencySOL, CurrencyUSDT, CurrencyUSDC}

type ProviderUser struct {
	ID    string
	Name  string
	Email string
}
