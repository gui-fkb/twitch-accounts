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
	//fmt.Scanln(&username)
	username = "guirerume_"

	userId, err := getUserId(username)
	if err != nil {
		log.Fatal(err)
	}

	if userId == "" {
		log.Fatal("User not found. Program exited.")
	} else {
		fmt.Println("User found with ID: " + userId)
	}

	myTestOauth := "l3l330k8ru88koihh11gjwg42dnc7l" // Letting this oauth token here for testing purposes. Just feel free to replace it with your own oauth token if it's not working.

	followTwitchUser(userId, myTestOauth)
}

func followTwitchUser(userid string, oauth string) {

}

func startFollowRequest(userId string, oauth string, XDeviceId string, ClientVersion string, ClientSessionId string, ClientIntegrity string, UserAgent string) {
	query := []shared.TwitchOperationQuery{
		{
			OperationName: "FollowButton_FollowUser",
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"disableNotifications": false,
					"targetID":             "${" + userId + "}",
				},
			},
			Extensions: map[string]interface{}{
				"persistedQuery": map[string]interface{}{
					"version":    1,
					"sha256Hash": "800e7346bdf7e5278a3c1d3f21b2b56e2639928f86815677a7126b093b2fdd08",
				},
			},
		},
	}

	queryBytes, _ := json.Marshal(query)

	req, _ := http.NewRequest("POST", "https://gql.twitch.tv/gql#origin=twilight", bytes.NewBuffer(queryBytes))
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Referer", "https://www.twitch.tv/")
	req.Header.Set("Client-Id", shared.Config.TwitchClientID)
	req.Header.Set("X-Device-Id", XDeviceId)
	req.Header.Set("Client-Version", ClientVersion)
	req.Header.Set("Client-Session", ClientSessionId)
	req.Header.Set("Authorization", "OAuth "+oauth)
	req.Header.Set("Client-Integrity", ClientIntegrity)
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func getUserId(username string) (string, error) {
	query := shared.TwitchOperationQuery{
		OperationName: "ChannelShell",
		Variables:     map[string]interface{}{"login": username},
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
