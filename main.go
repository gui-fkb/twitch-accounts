package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
)

func main() {
	fmt.Println("twitch-accounts by xBadApple -  https://github.com/xBadApple")

	if config.CapSolverKey == "your_captcha_key" {
		fmt.Println("It looks like your captcha solver API token isn't configured yet. Change it in the config.go file and run again.")
		os.Exit(1)
	}

	createNewAccount()
}

func createNewAccount() {
	randomUsername := getRandomUsername() + "_" + generateRandomID(3)

	fmt.Println(randomUsername)
}

func getRandomUsername() string {
	nameGenerator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

	name := strings.Replace(nameGenerator.Generate(), "-", "", -1)
	return name
}

func generateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	bytes := make([]byte, length)
	for i := range bytes {
		index := rand.Intn(length)
		bytes[i] = charset[index]
	}
	return string(bytes)
}
