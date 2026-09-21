package auth

func GetClaims(authHeader string, secret string, fallbackSecrets ...string) (*Claims, error) {

	token, err := ExtractBearer(authHeader)
	if err != nil {
		return nil, err
	}

	return ValidateJWT(token, secret, fallbackSecrets...)
}
