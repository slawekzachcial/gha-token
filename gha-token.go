package main

import (
	"flag"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	// apiUrlPtr := flag.String("apiUrl", "https://api.github.com", "GitHub API URL")
	appIdPtr := flag.Int("app", 0, "Appliction ID as defined in app settings")
	keyPathPtr := flag.String("keyPath", "", "Path to key PEM file generated in app settings")

	flag.Parse()
	// fmt.Printf("API: %s, App ID: %d, Key: %s\n", *apiUrlPtr, *appIdPtr, *keyPathPtr)

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

	tokenString, err := token.SignedString(signKey)

	fmt.Println(tokenString)
}

func handleErrorIfAny(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
}
