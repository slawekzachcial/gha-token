package main

import (
	"testing"

	jwt "github.com/golang-jwt/jwt"
)

// TODO: get these values from environment variables and use defaults below if empty
const testKeyPath = "gha-token-test.private-key.pem"
const testAppId = "148759"

func TestGetJWTTokenGenerated(t *testing.T) {
	jwtToken, err := getJwtToken(testAppId, testKeyPath)
	if err != nil {
		t.Error("JWT token generation failed")
	}
	if jwtToken == "" {
		t.Error("Non-empty JWT token expected")
	}
}

func TestGetJWTTokenWrongPath(t *testing.T) {
	_, err := getJwtToken(testAppId, "i_dont_exist.pem")
	if err == nil {
		t.Error("JWT token generation expected to fail")
	}
}

func TestGetJWTTokenAppIdInClaims(t *testing.T) {
	tokenString, _ := getJwtToken(testAppId, testKeyPath)

	token, _ := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("xxx"), nil
	})

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
		if claims.Issuer != testAppId {
			t.Errorf("Expected Issuer in the token '%s' was: %s but got: %s", tokenString, testAppId, claims.Issuer)
		}
	} else {
		t.Errorf("Unable to parse token: %s", tokenString)
	}
}
