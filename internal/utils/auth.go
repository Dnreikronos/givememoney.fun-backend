package utils

import (
	"errors"
	"strings"
)

func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return bearerToken[1], nil
}
