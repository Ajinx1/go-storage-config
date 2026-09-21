package auth

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func parseJWT(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(
				"unexpected signing method: %v",
				token.Header["alg"],
			)
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func ValidateJWT(tokenString string, secret string, fallbackSecrets ...string) (*Claims, error) {
	claims, err := parseJWT(tokenString, secret)
	if err == nil {
		normalizeClaims(claims)
		return claims, nil
	}

	for _, fallback := range fallbackSecrets {
		if fallback == "" || fallback == secret {
			continue
		}
		if c, fallbackErr := parseJWT(tokenString, fallback); fallbackErr == nil {
			normalizeClaims(c)
			return c, nil
		}
	}

	return nil, err
}

func normalizeClaims(claims *Claims) {
	if claims == nil {
		return
	}

	if claims.UserID == "" && string(claims.Sub) != "" {
		claims.UserID = string(claims.Sub)
	}

	if claims.IsGBOUser() {
		if claims.TaxID == "" && claims.EntityTaxID != "" {
			claims.TaxID = claims.EntityTaxID
		}
		if claims.TIN == "" && claims.EntityTaxID != "" {
			claims.TIN = claims.EntityTaxID
		}
		if claims.FullName == "" && claims.Username != "" {
			claims.FullName = claims.Username
		}
		if claims.Email == "" && claims.Username != "" && strings.Contains(claims.Username, "@") {
			claims.Email = claims.Username
		}
		if claims.ID == "" && claims.UserID != "" {
			claims.ID = claims.UserID
		}
	}
}
