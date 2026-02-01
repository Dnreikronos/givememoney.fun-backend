package validator

import (
	"strings"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/errors"
	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

const (
	base58Chars = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	ethAddrLen  = 40
	phantomMin  = 32
	phantomMax  = 44
)

// NormalizeAndValidateWalletAddress validates and normalizes the wallet address
// for the given provider. Returns the normalized address or an AppError.
func NormalizeAndValidateWalletAddress(provider utils.WalletProvider, address string) (string, *errors.AppError) {
	addr := strings.TrimSpace(address)
	providerLower := strings.ToLower(string(provider))

	switch providerLower {
	case "metamask":
		if strings.HasPrefix(strings.ToLower(addr), "0x") {
			addr = addr[2:]
		}
		if len(addr) != ethAddrLen {
			return "", errors.NewAppError(errors.ErrorCodeInvalidInput,
				"MetaMask wallet address must be exactly 40 hexadecimal characters",
				nil).WithContext("received_length", len(address)).WithContext("normalized_length", len(addr)).
				WithContext("hint", "Ethereum addresses are 40 hex chars (42 with 0x prefix)")
		}
		for _, char := range addr {
			if !isHexChar(char) {
				return "", errors.NewAppError(errors.ErrorCodeInvalidInput,
					"MetaMask wallet address must contain only hexadecimal characters (0-9, a-f, A-F)", nil)
			}
		}
		return strings.ToLower(addr), nil
	case "phantom":
		if len(addr) < phantomMin || len(addr) > phantomMax {
			return "", errors.NewAppError(errors.ErrorCodeInvalidInput,
				"Phantom wallet address must be between 32 and 44 characters", nil).
				WithContext("received_length", len(addr)).
				WithContext("hint", "Solana addresses are Base58 encoded strings, typically 32-44 characters")
		}
		for _, char := range addr {
			if !strings.ContainsRune(base58Chars, char) {
				return "", errors.NewAppError(errors.ErrorCodeInvalidInput,
					"Phantom wallet address must contain only valid Base58 characters (excludes 0, O, I, l)", nil)
			}
		}
		return addr, nil
	default:
		return "", errors.NewAppError(errors.ErrorCodeInvalidInput, "Invalid provider", nil).
			WithContext("supported", []string{"metamask", "phantom"}).
			WithContext("received_provider", provider)
	}
}

func isHexChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
