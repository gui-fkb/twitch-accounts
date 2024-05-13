package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"twitch-accounts/shared"
)

func main() {
	fmt.Println("twitch-accounts by xBadApple -  https://github.com/xBadApple")
	//shared.FastEmailTest() // Uncomment this line if you want to test the trash email in a fast way, dont forget to enable breakpoints inside the function

	if shared.Config.CapSolverKey == "your_captcha_key" {
		log.Fatal("It looks like your captcha solver API token isn't configured yet. Change it in the shared.Config.go file and run again.")
	}

	var username string
	fmt.Println("Enter the username you want to follow: ")
	fmt.Scanln(&username)

	userId, err := getUserId(username)
	if err != nil {
		log.Fatal(err)
	}

	if userId == "" {
		log.Fatal("User not found. Program exited.")
	} else {
		fmt.Println("User found with ID: " + userId)
	}

}

func followTwitchUser(userid string, oauth string) {

}

func getUserId(username string) (string, error) {
	query := shared.QueryUserId{
		OperationName: "ChannelShell",
		Variables:     map[string]string{"login": username},
		Extensions: map[string]interface{}{
			"persistedQuery": map[string]interface{}{
				"version":    1,
				"sha256Hash": "580ab410bcd0c1ad194224957ae2241e5d252b2c5173d8e0cce9d32d5bb14efe",
			},
		},
	}

	queryBytes, _ := json.Marshal(query)
	req, _ := http.NewRequest("POST", "https://gql.twitch.tv/gql", bytes.NewBuffer(queryBytes))
	req.Header.Set("Client-ID", shared.Config.TwitchClientID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	data := result["data"].(map[string]interface{})
	userOrError := data["userOrError"].(map[string]interface{})

	if userOrError != nil && userOrError["id"] != nil {
		return userOrError["id"].(string), nil
	} else {
		return "", nil
	}
}
