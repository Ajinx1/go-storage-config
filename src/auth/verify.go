package auth

func GetClaims(authHeader string, secret string) (*Claims, error) {

	token, err := ExtractBearer(authHeader)
	if err != nil {
		return nil, err
	}

	return ValidateJWT(token, secret)
}
