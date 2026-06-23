package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
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
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	p := jwt.NewParser()
	type ClaimStruct struct {
		//Id uuid.UUID `json:"id"`
		jwt.RegisteredClaims
	}

	// ParseWithClaims name is a litte misleading, it parses /into/ the claims struct, not from.
	token, err := p.ParseWithClaims(tokenString, &ClaimStruct{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.UUID{}, err
	} else if claims, ok := token.Claims.(*ClaimStruct); ok {
		user_id, tokenErr := uuid.Parse(claims.Subject)
		if tokenErr != nil {
			return uuid.UUID{}, tokenErr
		} else {
			return user_id, nil
		}

	} else {
		return uuid.UUID{}, errors.New("Unknown claims type, cannot proceed")
	}

	return uuid.UUID{}, errors.New("Unknown error has occurred")
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := strings.Split(headers.Get("Authorization"), " ")[1]
	fmt.Printf("%v", tokenString)
	if tokenString == "" {
		return "", errors.New("Bearer not found in request headers")
	}

	return tokenString, nil
}

func MakeRefreshToken() string {
	b := [32]byte{}
	rand.Read(b)

	encodedStr := hex.EncodeToString(b)
	return encodedStr
}
