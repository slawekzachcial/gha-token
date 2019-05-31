package main

import (
	"encoding/json"
	"flag"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	logger "log"
	"net/http"
	"os"
	"strings"
	"time"
)

type installationToken struct {
	token     string `json:"token"`
	expiresAt string `json:"expires_at"`
}

type installation struct {
	id              int    `json:"id"`
	accessTokensURL string `json:"access_tokens_url"`
	repositoriesURL string `json:"repositories_url"`
}

type repository struct {
	name  string `json:"name"`
	owner struct {
		login string `json:"login"`
	} `json:"owner"`
}

type repositories struct {
	List []repository `json:"repositories"`
}

var verbose bool

func main() {
	apiURLPtr := flag.String("apiUrl", "https://api.github.com", "GitHub API URL")
	appIDPtr := flag.String("app", "", "Appliction ID as defined in app settings. Required")
	keyPathPtr := flag.String("keyPath", "", "Path to key PEM file generated in app settings. Required")
	installIDPtr := flag.String("inst", "", "Installation ID of the application")
	repoPtr := flag.String("repo", "", "{owner/repo} of the GitHub repository")
	flag.BoolVar(&verbose, "v", false, "Verbose stderr")

	flag.Parse()

	if *appIDPtr == "" || *keyPathPtr == "" {
		usage()
	}

	jwtToken := getJwtToken(*appIDPtr, *keyPathPtr)

	var token string

	if *installIDPtr == "" && *repoPtr == "" {
		log("Generated JWT token for app ID %s\n", *appIDPtr)
		token = jwtToken
	} else if *installIDPtr != "" {
		installToken := getInstallationToken(*apiURLPtr, jwtToken, *appIDPtr, *installIDPtr)
		log("Generated installation token for app ID %s and installation ID %s that expires at %s\n", *appIDPtr, *installIDPtr, installToken.expiresAt)
		token = installToken.token
	} else {
		repoInfo := strings.Split(*repoPtr, "/")
		if len(repoInfo) != 2 {
			logger.Fatalln("-repo argument value must be owner/repo but was: %s", *repoPtr)
		}
		owner, repo := repoInfo[0], repoInfo[1]
		installToken, err := getInstallationTokenForRepo(*apiURLPtr, jwtToken, *appIDPtr, owner, repo)
		handleErrorIfAny(err)
		log("Generated installation token for app ID %s and repo %s that expires at %s\n", *appIDPtr, *repoPtr, installToken.expiresAt)
		token = installToken.token
	}

	fmt.Print(token)
}

func getJwtToken(appID string, keyPath string) string {
	keyBytes, err := ioutil.ReadFile(keyPath)
	handleErrorIfAny(err)

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	handleErrorIfAny(err)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 10).Unix(),
		"iss": appID,
	})

	jwtTokenString, err := token.SignedString(signKey)
	handleErrorIfAny(err)

	return jwtTokenString
}

func httpJSON(method string, url string, authorization string, result interface{}) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, nil)
	handleErrorIfAny(err)
	req.Header.Add("Authorization", authorization)
	req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

	log("GitHub request: %s", req)

	resp, err := client.Do(req)
	handleErrorIfAny(err)

	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)
	handleErrorIfAny(err)

	log("GitHub response: %s", respData)

	json.Unmarshal(respData, &result)
}

func getInstallationToken(apiURL string, jwtToken string, appID string, installationID string) installationToken {
	var installationToken installationToken
	httpJSON("POST", fmt.Sprintf("%s/app/installations/%s/access_tokens", apiURL, installationID), "Bearer "+jwtToken, &installationToken)

	return installationToken
}

func getInstallationTokenForRepo(apiURL string, jwtToken string, appID string, owner string, repo string) (installationToken, error) {
	var installations []installation
	httpJSON("GET", apiURL+"/app/installations", "Bearer "+jwtToken, &installations)

	for _, installation := range installations {
		var installationToken installationToken
		httpJSON("POST", installation.accessTokensURL, "Bearer "+jwtToken, &installationToken)

		var repos repositories
		httpJSON("GET", installation.repositoriesURL, "token "+installationToken.token, &repos)

		for _, repository := range repos.List {
			if owner == repository.owner.login && repo == repository.name {
				return installationToken, nil
			}
		}
	}
	var empty installationToken
	return empty, fmt.Errorf("Unable to find repository %s/%s in installations of app ID %s", owner, repo, appID)
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

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
