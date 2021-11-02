package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	logger "log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt"
	flag "github.com/spf13/pflag"
)

const githubApiUrl = "https://api.github.com"

type installationToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type installation struct {
	ID              int    `json:"id"`
	AccessTokensURL string `json:"access_tokens_url"`
}

type config struct {
	apiURL    string
	appID     string
	keyPath   string
	installID string
	repoOwner string
	repoName  string
}

var verbose bool

func main() {
	var cfg = parseFlags()

	jwtToken, err := getJwtToken(cfg.appID, cfg.keyPath)
	handleErrorIfAny(err)

	var token string

	if cfg.installID == "" && cfg.repoOwner == "" {
		log("Generated JWT token for app ID %s\n", cfg.appID)
		token = jwtToken
	} else if cfg.installID != "" {
		installToken, err := getInstallationToken(cfg.apiURL, jwtToken, cfg.appID, cfg.installID)
		handleErrorIfAny(err)
		log("Generated installation token for app ID %s and installation ID %s that expires at %s\n", cfg.appID, cfg.installID, installToken.ExpiresAt)
		token = installToken.Token
	} else {
		installToken, err := getInstallationTokenForRepo(cfg.apiURL, jwtToken, cfg.appID, cfg.repoOwner, cfg.repoName)
		handleErrorIfAny(err)
		log("Generated installation token for app ID %s and repo %s/%s that expires at %s\n", cfg.appID, cfg.repoOwner, cfg.repoName, installToken.ExpiresAt)
		token = installToken.Token
	}

	fmt.Print(token)
}

func parseFlags() config {
	var cfg config

	flag.StringVarP(&cfg.apiURL, "apiUrl", "g", githubApiUrl, "GitHub API URL")
	flag.StringVarP(&cfg.appID, "appId", "a", "", "Application ID as defined in app settings (Required)")
	flag.StringVarP(&cfg.keyPath, "keyPath", "k", "", "Path to key PEM file generated in app settings (Required)")
	flag.StringVarP(&cfg.installID, "installId", "i", "", "Installation ID of the application")
	repoPtr := flag.StringP("repo", "r", "", "{owner/repo} of the GitHub repository")
	flag.BoolVarP(&verbose, "verbose", "v", false, "Verbose stderr")

	flag.Parse()

	if len(os.Args) == 1 {
		usage("")
	}

	if cfg.appID == "" {
		usage("appId is required")
	}

	if cfg.keyPath == "" {
		usage("keyPath is required")
	}

	if *repoPtr != "" {
		repoInfo := strings.Split(*repoPtr, "/")
		if len(repoInfo) != 2 {
			usage("repo argument value must be owner/repo but was: " + *repoPtr)
		}
		cfg.repoOwner, cfg.repoName = repoInfo[0], repoInfo[1]
	}

	return cfg
}

func getJwtToken(appID string, keyPath string) (string, error) {
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return "", err
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return "", err
	}

	now := time.Now()
	// StandardClaims: https://pkg.go.dev/github.com/golang-jwt/jwt#StandardClaims
	// Issuer: iss, IssuedAt: iat, ExpiresAt: exp
	claims := &jwt.StandardClaims{
		Issuer:    appID,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(time.Minute * 10).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	jwtTokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}

	return jwtTokenString, nil
}

func logRequest(req *http.Request) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		log("GitHub request:\n%s", string(reqDump))
	} else {
		log("Unable to log GitHub request: %s", err)
	}
}

func logResponse(resp *http.Response) {
	respDump, err := httputil.DumpResponse(resp, true)
	if err == nil {
		log("GitHub response:\n%s", string(respDump))
	} else {
		log("Unable to log GitHub response: %s", err)
	}
}

// TODO: return result instead of passing it as param
func httpJSON(method string, url string, authorization string, result interface{}) error {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", authorization)
	req.Header.Add("Accept", "application/vnd.github.v3+json")

	logRequest(req)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	logResponse(resp)

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Expected HTTP status code 2xx but got %d", resp.StatusCode)
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal(respData, &result)

	log("%s", result)

	return nil
}

func getInstallationToken(apiURL string, jwtToken string, appID string, installationID string) (installationToken, error) {
	var token installationToken
	err := httpJSON("POST", fmt.Sprintf("%s/app/installations/%s/access_tokens", apiURL, installationID), "Bearer "+jwtToken, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}

func getInstallationTokenForRepo(apiURL string, jwtToken string, appID string, owner string, repo string) (installationToken, error) {
	var repoInstallation installation
	var token installationToken

	err := httpJSON("GET", fmt.Sprintf("%s/repos/%s/%s/installation", apiURL, owner, repo), "Bearer "+jwtToken, &repoInstallation)
	if err != nil {
		return token, err
	}

	err = httpJSON("POST", repoInstallation.AccessTokensURL, "Bearer "+jwtToken, &token)
	if err != nil {
		return token, err
	}

	return token, nil
}

func log(format string, v ...interface{}) {
	if verbose {
		logger.Printf(format, v...)
	}
}

func handleErrorIfAny(err error) {
	if err != nil {
		logger.Fatalln(err)
	}
}

func usage(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", msg)
	}
	fmt.Fprintf(os.Stderr, "Usage: gha-token [flags]\n\nFlags:\n")
	flag.PrintDefaults()
	os.Exit(1)
}
