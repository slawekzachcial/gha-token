package main

import (
	"encoding/json"
	"flag"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	logger "log"
	"net/http"
	"strings"
	"time"
)

type InstallationToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type Installation struct {
	Id              int    `json:"id"`
	AccessTokensUrl string `json:"access_tokens_url"`
	RepositoriesUrl string `json:"repositories_url"`
}

type Repository struct {
	Name  string `json:"name"`
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
}

type Repositories struct {
	List []Repository `json:"repositories"`
}

var verbose bool

func main() {
	apiUrlPtr := flag.String("apiUrl", "https://api.github.com", "GitHub API URL")
	appIdPtr := flag.Int("app", 0, "Appliction ID as defined in app settings")
	keyPathPtr := flag.String("keyPath", "", "Path to key PEM file generated in app settings")
	installIdPtr := flag.Int("inst", -1, "Installation ID of the application")
	repoPtr := flag.String("repo", "", "{owner/repo} of the GitHub repository")
	flag.BoolVar(&verbose, "v", false, "Verbose stderr")

	flag.Parse()
	log("API: %s, App ID: %d, Key: %s\n", *apiUrlPtr, *appIdPtr, *keyPathPtr)

	jwtToken := jwtToken(*appIdPtr, *keyPathPtr)

	var token string

	if *installIdPtr == -1 && *repoPtr == "" {
		log("Generated JWT token for app ID %d\n", *appIdPtr)
		token = jwtToken
	} else if *installIdPtr != -1 {
		installationToken := installationToken(*apiUrlPtr, jwtToken, *appIdPtr, *installIdPtr)
		log("Generated installation token for app ID %d and installation ID %d that expires at %s\n", *appIdPtr, *installIdPtr, installationToken.ExpiresAt)
		token = installationToken.Token
	} else {
		repoInfo := strings.Split(*repoPtr, "/")
		if len(repoInfo) != 2 {
			logger.Fatalln("-repo argument value must be owner/repo but was: %s", *repoPtr)
		}
		owner, repo := repoInfo[0], repoInfo[1]
		installationToken, err := installationTokenForRepo(*apiUrlPtr, jwtToken, *appIdPtr, owner, repo)
		handleErrorIfAny(err)
		log("Generated installation token for app ID %d and repo %s that expires at %s\n", *appIdPtr, *repoPtr, installationToken.ExpiresAt)
		token = installationToken.Token
	}

	fmt.Print(token)
}

func jwtToken(appId int, keyPath string) string {
	keyBytes, err := ioutil.ReadFile(keyPath)
	handleErrorIfAny(err)

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	handleErrorIfAny(err)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 10).Unix(),
		"iss": appId,
	})

	jwtTokenString, err := token.SignedString(signKey)
	handleErrorIfAny(err)

	return jwtTokenString
}

func httpJson(method string, url string, authorization string, result interface{}) {
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

func installationToken(apiUrl string, jwtToken string, appId int, installationId int) InstallationToken {
	var installationToken InstallationToken
	httpJson("POST", fmt.Sprintf("%s/app/installations/%d/access_tokens", apiUrl, installationId), "Bearer "+jwtToken, &installationToken)

	return installationToken
}

func installationTokenForRepo(apiUrl string, jwtToken string, appId int, owner string, repo string) (InstallationToken, error) {
	var installations []Installation
	httpJson("GET", apiUrl+"/app/installations", "Bearer "+jwtToken, &installations)

	for _, installation := range installations {
		var installationToken InstallationToken
		httpJson("POST", installation.AccessTokensUrl, "Bearer "+jwtToken, &installationToken)

		var repositories Repositories
		httpJson("GET", installation.RepositoriesUrl, "token "+installationToken.Token, &repositories)

		for _, repository := range repositories.List {
			if owner == repository.Owner.Login && repo == repository.Name {
				return installationToken, nil
			}
		}
	}
	var empty InstallationToken
	return empty, fmt.Errorf("Unable to find repository %s/%s in installations of app ID %d", owner, repo, appId)
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
