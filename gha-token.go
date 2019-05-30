package main

import (
	"encoding/json"
	"flag"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	logger "log"
	"net/http"
	"time"
)

type InstallationToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

var verbose bool

func main() {
	apiUrlPtr := flag.String("apiUrl", "https://api.github.com", "GitHub API URL")
	appIdPtr := flag.Int("app", 0, "Appliction ID as defined in app settings")
	keyPathPtr := flag.String("keyPath", "", "Path to key PEM file generated in app settings")
	installIdPtr := flag.Int("inst", -1, "Installation ID of the application")
	flag.BoolVar(&verbose, "v", false, "Verbose stderr")

	flag.Parse()
	log("API: %s, App ID: %d, Key: %s\n", *apiUrlPtr, *appIdPtr, *keyPathPtr)

	keyBytes, err := ioutil.ReadFile(*keyPathPtr)
	handleErrorIfAny(err)

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	handleErrorIfAny(err)

	// fmt.Println(signKey)

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(time.Minute * 10).Unix(),
		"iss": *appIdPtr,
	})

	jwtTokenString, err := token.SignedString(signKey)
	handleErrorIfAny(err)

	if *installIdPtr == -1 {
		log("Generated JWT token for app ID %d\n", *appIdPtr)
		fmt.Print(jwtTokenString)
	} else {
		client := &http.Client{}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/app/installations/%d/access_tokens", *apiUrlPtr, *installIdPtr), nil)
		handleErrorIfAny(err)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", jwtTokenString))
		req.Header.Add("Accept", "application/vnd.github.machine-man-preview+json")

		resp, err := client.Do(req)
		handleErrorIfAny(err)

		defer resp.Body.Close()

		respData, err := ioutil.ReadAll(resp.Body)
		handleErrorIfAny(err)

		var installationToken InstallationToken
		json.Unmarshal(respData, &installationToken)

		log("Generated installation token for app ID %d and installation ID %d that expires at %s\n", *appIdPtr, *installIdPtr, installationToken.ExpiresAt)
		fmt.Print(installationToken.Token)
	}
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
