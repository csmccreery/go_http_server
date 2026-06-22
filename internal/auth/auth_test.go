package auth

import (
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	_, err := HashPassword("TestPassword")
	if err != nil {
		t.Errorf("HashPassword('TestPassword'); failed")
	}
}

func TestCheckHashPassword(t *testing.T) {
	password := "TestPassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword('TestPassword'); failed")
	}

	res, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("CheckHashPassword; failed")
	}

	if !res {
		t.Errorf("CheckHashPassword failed to produce unique hash")
	}
}

func TestMakeJWT(t *testing.T) {
	secret := "password"
	uid := uuid.New()
	expiresIn, err := time.ParseDuration("10m")
	if err != nil {
		t.Errorf("MakeJWT; failed to make JWT")
	}

	_, err = MakeJWT(uid, secret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT; failed to make JWT")
	}
}

func TestValidateJWT(t *testing.T) {
	ss, err := MakeJWT(uuid.New(), "password", time.Duration(1000000000000))
	if err != nil {
		t.Errorf("MakeJWT; Failed to create JWT")
	}

	_, err = ValidateJWT(ss, "password")
	if err != nil {
		fmt.Printf("Error: %v", err)
		t.Errorf("ValidateJWT; Failed to validate JWT")
	}

}
