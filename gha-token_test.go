package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	jwt "github.com/golang-jwt/jwt"
)

// TODO: get these values from environment variables and use defaults below if empty
const testKeyPath = "gha-token-test.private-key.pem"
const testAppId = "148759"
const testAppInstallId = "20435383"
const testAppInstallRepo = "slawekzachcial/gha-token-test"

// Flag that can be used to print verbose output from gha-token:
// go unit [some args] -args -ghaTokenVerbose=true
var ghaTokenVerbose = flag.Bool("ghaTokenVerbose", false, "Show verbose output for gha-token")

func TestMain(m *testing.M) {
	flag.Parse()
	verbose = *ghaTokenVerbose
	os.Exit(m.Run())
}

func TestGetInstallationTokenForInstallId(t *testing.T) {
	jwtToken, err := getJwtToken(testAppId, testKeyPath)
	if err != nil {
		t.Fatalf("Error getting JWT token: %v", err)
	}

	token, err := getInstallationToken(githubApiUrl, jwtToken, testAppId, testAppInstallId)
	if err != nil {
		t.Fatalf("Error getting installation token: %v", err)
	}
	if token.Token == "" {
		t.Error("Non-empty installation token expected")
	}

	type Repo struct {
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
	}

	var repo Repo

	err = httpJSON("GET", fmt.Sprintf("%s/repos/%s", githubApiUrl, testAppInstallRepo), "token "+token.Token, &repo)
	if err != nil {
		t.Fatalf("Error using installation token: %v", err)
	}

	if repo.FullName != testAppInstallRepo {
		t.Errorf("Expected repo full name: %s but was: %s", testAppInstallRepo, repo.FullName)
	}
	if !repo.Private {
		t.Error("Expected repo to be private")
	}
}

func TestGetInstallationTokenForBadInstallId(t *testing.T) {
	jwtToken, err := getJwtToken(testAppId, testKeyPath)
	if err != nil {
		t.Fatalf("Error getting JWT token: %v", err)
	}

	_, err = getInstallationToken(githubApiUrl, jwtToken, testAppId, "00000")
	if err == nil {
		t.Error("Installation token retrieval expected to fail")
	}
}

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
