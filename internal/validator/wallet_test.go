package validator

import (
	"strings"
	"testing"

	"github.com/Dnreikronos/givememoney.fun-backend/internal/utils"
)

func TestNormalizeAndValidateWalletAddress(t *testing.T) {
	tests := []struct {
		name        string
		provider    utils.WalletProvider
		address     string
		wantNorm    string
		wantErr     bool
		wantMsgPart string
	}{
		{
			name:        "metamask valid without 0x",
			provider:    utils.MetamaskWalletProvider,
			address:     "a1b2c3d4e5f6789012345678901234567890abcd",
			wantNorm:    "a1b2c3d4e5f6789012345678901234567890abcd",
			wantErr:     false,
			wantMsgPart: "",
		},
		{
			name:        "metamask valid with 0x",
			provider:    utils.MetamaskWalletProvider,
			address:     "0xAbCdEf1234567890AbCdEf1234567890AbCdEf12",
			wantNorm:    "abcdef1234567890abcdef1234567890abcdef12",
			wantErr:     false,
			wantMsgPart: "",
		},
		{
			name:        "metamask wrong length",
			provider:    utils.MetamaskWalletProvider,
			address:     "short",
			wantNorm:    "",
			wantErr:     true,
			wantMsgPart: "40 hexadecimal",
		},
		{
			name:        "metamask non-hex",
			provider:    utils.MetamaskWalletProvider,
			address:     "g1b2c3d4e5f6789012345678901234567890abcd",
			wantNorm:    "",
			wantErr:     true,
			wantMsgPart: "hexadecimal",
		},
		{
			name:        "phantom valid",
			provider:    utils.PhantomWalletProvider,
			address:     "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU",
			wantNorm:    "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU",
			wantErr:     false,
			wantMsgPart: "",
		},
		{
			name:        "phantom too short",
			provider:    utils.PhantomWalletProvider,
			address:     "short",
			wantNorm:    "",
			wantErr:     true,
			wantMsgPart: "32 and 44",
		},
		{
			name:        "invalid provider",
			provider:    utils.WalletProvider("unknown"),
			address:     "anything",
			wantNorm:    "",
			wantErr:     true,
			wantMsgPart: "Invalid provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNorm, gotErr := NormalizeAndValidateWalletAddress(tt.provider, tt.address)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("NormalizeAndValidateWalletAddress() err = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
			if !tt.wantErr && gotNorm != tt.wantNorm {
				t.Errorf("NormalizeAndValidateWalletAddress() norm = %q, want %q", gotNorm, tt.wantNorm)
			}
			if tt.wantErr && tt.wantMsgPart != "" && gotErr != nil && !strings.Contains(gotErr.Message, tt.wantMsgPart) {
				t.Errorf("NormalizeAndValidateWalletAddress() message = %q, want to contain %q", gotErr.Message, tt.wantMsgPart)
			}
		})
	}
}

