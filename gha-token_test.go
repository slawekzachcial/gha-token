package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	jwt "github.com/golang-jwt/jwt"
)

var testConfig config

// Flag that can be used to print verbose output from gha-token:
// go unit [some args] -args -ghaTokenVerbose=true
var ghaTokenVerbose = flag.Bool("ghaTokenVerbose", false, "Show verbose output for gha-token")

func TestMain(m *testing.M) {
	testConfig.apiURL = getenv("TEST_GHA_TOKEN_API_URL", githubApiUrl)
	testConfig.appID = getenv("TEST_GHA_TOKEN_APP_ID", "148759")
	testConfig.keyPath = getenv("TEST_GHA_TOKEN_KEY_PATH", "gha-token-test.private-key.pem")
	testConfig.installID = getenv("TEST_GHA_TOKEN_APP_INSTALL_ID", "20435383")
	testConfig.repoOwner = getenv("TEST_GHA_TOKEN_APP_INSTALL_REPO_OWNER", "slawekzachcial")
	testConfig.repoName = getenv("TEST_GHA_TOKEN_APP_INSTALL_REPO_NAME", "gha-token-test")

	flag.Parse()
	verbose = *ghaTokenVerbose
	os.Exit(m.Run())
}

func getenv(name, defaultValue string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return defaultValue
}

func useInstallationToken(t *testing.T, token string) {
	type Repo struct {
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
	}

	var repo Repo

	err := httpJSON("GET", fmt.Sprintf("%s/repos/%s/%s", testConfig.apiURL, testConfig.repoOwner, testConfig.repoName), "token "+token, &repo)
	if err != nil {
		t.Fatalf("Error using installation token: %v", err)
	}

	if repo.FullName != testConfig.repoOwner+"/"+testConfig.repoName {
		t.Errorf("Expected repo full name: %s/%s but was: %s", testConfig.repoOwner, testConfig.repoName, repo.FullName)
	}
	if !repo.Private {
		t.Error("Expected repo to be private")
	}
}

func TestGetInstallationTokenForRepo(t *testing.T) {
	jwtToken, err := getJwtToken(testConfig.appID, testConfig.keyPath, jwtExpirationSecs)
	if err != nil {
		t.Fatalf("Error getting JWT token: %v", err)
	}

	token, err := getInstallationTokenForRepo(githubApiUrl, jwtToken, testConfig.appID, testConfig.repoOwner, testConfig.repoName)
	if err != nil {
		t.Fatalf("Error getting installation token: %v", err)
	}
	if token.Token == "" {
		t.Error("Non-empty installation token expected")
	}

	useInstallationToken(t, token.Token)
}

func TestGetInstallationTokenForBadRepo(t *testing.T) {
	jwtToken, err := getJwtToken(testConfig.appID, testConfig.keyPath, jwtExpirationSecs)
	if err != nil {
		t.Fatalf("Error getting JWT token: %v", err)
	}

	_, err = getInstallationTokenForRepo(githubApiUrl, jwtToken, testConfig.appID, "bad", "repo0")
	if err == nil {
		t.Error("Installation token retrieval expected to fail")
	}
}

func TestGetInstallationTokenForInstallId(t *testing.T) {
	jwtToken, err := getJwtToken(testConfig.appID, testConfig.keyPath, jwtExpirationSecs)
	if err != nil {
		t.Fatalf("Error getting JWT token: %v", err)
	}

	token, err := getInstallationToken(githubApiUrl, jwtToken, testConfig.appID, testConfig.installID)
	if err != nil {
		t.Fatalf("Error getting installation token: %v", err)
	}
	if token.Token == "" {
		t.Error("Non-empty installation token expected")
	}

	useInstallationToken(t, token.Token)
}

func TestGetInstallationTokenForBadInstallId(t *testing.T) {
	jwtToken, err := getJwtToken(testConfig.appID, testConfig.keyPath, jwtExpirationSecs)
	if err != nil {
		t.Fatalf("Error getting JWT token: %v", err)
	}

	_, err = getInstallationToken(githubApiUrl, jwtToken, testConfig.appID, "00000")
	if err == nil {
		t.Error("Installation token retrieval expected to fail")
	}
}

func TestGetJWTTokenGenerated(t *testing.T) {
	jwtToken, err := getJwtToken(testConfig.appID, testConfig.keyPath, jwtExpirationSecs)
	if err != nil {
		t.Error("JWT token generation failed")
	}
	if jwtToken == "" {
		t.Error("Non-empty JWT token expected")
	}
}

func TestGetJWTTokenWrongPath(t *testing.T) {
	_, err := getJwtToken(testConfig.appID, "i_dont_exist.pem", jwtExpirationSecs)
	if err == nil {
		t.Error("JWT token generation expected to fail")
	}
}

func TestGetJWTTokenAppIdInClaims(t *testing.T) {
	tokenString, _ := getJwtToken(testConfig.appID, testConfig.keyPath, jwtExpirationSecs)

	token, _ := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("xxx"), nil
	})

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
		if claims.Issuer != testConfig.appID {
			t.Errorf("Expected Issuer in the token '%s' was: %s but got: %s", tokenString, testConfig.appID, claims.Issuer)
		}
	} else {
		t.Errorf("Unable to parse token: %s", tokenString)
	}
}

func TestGetJWTTokenWithCustomExpiration(t *testing.T) {
	customExpirationSecs := 3600
	tokenString, _ := getJwtToken(testConfig.appID, testConfig.keyPath, customExpirationSecs)

	token, _ := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("xxx"), nil
	})

	if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
		tokenExpirationSecs := claims.ExpiresAt - claims.IssuedAt
		if int(tokenExpirationSecs) != customExpirationSecs {
			t.Errorf("Expected JWT token custom expiration to be: %ds but was: %ds", customExpirationSecs, tokenExpirationSecs)
		}
	} else {
		t.Errorf("Unable to parse token: %s", tokenString)
	}
}
