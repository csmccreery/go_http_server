package auth

import (
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password string, hash string) (bool, error) {
	valid, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil || !valid {
		return false, err
	}

	return true, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	claims := &jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

func validate(token *jwt.Token) []byte {
	//
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	p := jwt.NewParser()
	token, err := p.ParseWithClaims(tokenString, claims, validate)
}
