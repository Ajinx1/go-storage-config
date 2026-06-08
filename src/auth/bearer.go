package auth

import (
	"errors"
	"strings"
)

func ExtractBearer(authHeader string) (string, error) {

	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	parts := strings.Split(authHeader, " ")

	if len(parts) != 2 {
		return "", errors.New("invalid bearer token")
	}

	if !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid bearer token")
	}

	return parts[1], nil
}
