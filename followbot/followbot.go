package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"twitch-accounts/shared"
)

var friendRequestsSent = 0
var tokensFile string = "./results/tokens.txt"

func main() {
	fmt.Println("twitch-accounts by xBadApple -  https://github.com/xBadApple")
	//shared.FastEmailTest() // Uncomment this line if you want to test the trash email in a fast way, dont forget to enable breakpoints inside the function

	if shared.Config.CapSolverKey == "your_captcha_key" {
		log.Fatal("It looks like your captcha solver API token isn't configured yet. Change it in the shared.Config.go file and run again.")
	}

	var username string
	fmt.Println("Enter the username you want to follow: ")
	//username = "guirerume_" // Take this opportunity to follow me :)
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

	tokens := getTokenList()
	//myTestOauth := "l3l330k8ru88koihh11gjwg42dnc7l" // Letting this oauth token here for testing purposes. Just feel free to replace it with your own oauth token if it's not working.

	var wg sync.WaitGroup
	sem := make(chan bool, 5) // Limit to 5 concurrent goroutines

	for _, oauthToken := range tokens {
		wg.Add(1)
		go func(oauthToken string) {
			oauthToken = strings.TrimSpace(oauthToken)
			sem <- true // Will block if there is already 5 goroutines running
			defer func() {
				<-sem // Release the slot
				wg.Done()
			}()

			fmt.Println("Using token: " + oauthToken)
			followTwitchUser(userId, oauthToken)
		}(oauthToken)
	}

	wg.Wait()
	close(sem)

	fmt.Println("All follow requests sent.")
}

func getTokenList() []string {
	file, err := os.Open(tokensFile)
	if err != nil {
		fmt.Println(err)
		return make([]string, 0)
	}
	defer file.Close()

	var tokens []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tokens = append(tokens, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		return make([]string, 0)
	}

	return tokens
}

func followTwitchUser(userid string, oauth string) {
	fmt.Println("Getting twitch cookies.")
	cookies, err := getTwitchCookies()
	if err != nil {
		fmt.Println(err, "\n account creation exited")
		return
	}

	for i := 0; i < 3; i++ {
		fmt.Println("Following attempt", i+1, " of ", 3)

		fmt.Println("Getting kasada code")
		taskResponse, err := kasadaResolver()
		if err != nil {
			fmt.Println(err, "\n account creation exited")
			return
		}

		clientSessionId := shared.GenerateRandomID(16)
		xDeviceId := cookies["unique_id"]
		clientVersion := "3040e141-5964-4d72-b67d-e73c1cf355b5"
		clientRequestId := shared.GenerateRandomID(32)

		fmt.Println("Getting public integrity token...")
		publicIntegrityData, err := publicIntegrityGetToken(xDeviceId, clientRequestId, clientSessionId, clientVersion, taskResponse.Solution["x-kpsdk-ct"], taskResponse.Solution["x-kpsdk-cd"], oauth, taskResponse.Solution["user-agent"])
		if err != nil {
			fmt.Println(err, "\n error getting public integrity token - account creation exited")
			continue
		}

		jsonResp := startFollowRequest(userid, oauth, xDeviceId, clientVersion, clientSessionId, publicIntegrityData.Token, taskResponse.Solution["user-agent"])
		if strings.Contains(jsonResp, "displayName") {
			friendRequestsSent++
			fmt.Println("Follow request sent succesfully. ", " - friend requests sent in total:", friendRequestsSent)
			break
		}
	}
}

func startFollowRequest(userId string, oauth string, XDeviceId string, ClientVersion string, ClientSessionId string, ClientIntegrity string, UserAgent string) string {
	query := []shared.TwitchOperationQuery{
		{
			OperationName: "FollowButton_FollowUser",
			Variables: map[string]interface{}{
				"input": map[string]interface{}{
					"disableNotifications": false,
					"targetID":             userId,
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
		return ""
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
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

func kasadaResolver() (*shared.ResultTaskResponse, error) {
	taskResponse, err := createKasadaTask()
	if err != nil {
		return nil, err
	}

	maxAttemps := 12
	for i := 0; i < maxAttemps; i++ {
		time.Sleep(time.Millisecond * 400)

		taskResult, err := getTaskResult(taskResponse.TaskId)
		if err != nil {
			return nil, err
		}

		if taskResult.Status == "ready" {
			return taskResult, nil
		}
	}

	return nil, errors.New("kasada task took too long to resolve")
}

func createKasadaTask() (*shared.CreateTaskResponse, error) {
	// There is not the need to use proxy here, because the kasada task is not being blocked by the server

	requestBody := shared.CreateKasadaTask{
		ApiKey: shared.Config.CapSolverKey,
		Task: shared.Task{
			Type:   "KasadaCaptchaSolver",
			Pjs:    "https://k.twitchcdn.net/149e9513-01fa-4fb0-aad4-566afd725d1b/2d206a39-8ed7-437e-a3be-862e0f06eea3/p.js",
			CdOnly: false,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://salamoonder.com/api/createTask", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	taskResp := &shared.CreateTaskResponse{}

	err = json.Unmarshal(body, taskResp)
	if err != nil {
		return nil, err
	}

	return taskResp, nil
}

func getTaskResult(taskId string) (*shared.ResultTaskResponse, error) {
	// There is not the need to use proxy here, because the kasada task is not being blocked by the server
	task := shared.GetTaskResult{TaskId: taskId}

	jsonBody, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("https://salamoonder.com/api/getTaskResult", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	taskResponse := &shared.ResultTaskResponse{}

	err = json.Unmarshal(body, &taskResponse)
	if err != nil {
		return nil, err
	}

	return taskResponse, nil
}

func publicIntegrityGetToken(XDeviceId, ClientRequestId, ClientSessionId, ClientVersion, kpsdkct, kpsdkcd, accesstoken, current_useragent string) (publicIntegrity *shared.PublicIntegrityData, err error) {
	requestBody := []byte("{}")

	headers := map[string]string{
		"User-Agent":        current_useragent,
		"Accept":            "application/json",
		"Accept-Language":   "en-US",
		"Accept-Encoding":   "identity",
		"Authorization":     "OAuth " + accesstoken,
		"Referer":           "https://www.twitch.tv/",
		"Client-Id":         shared.Config.TwitchClientID,
		"X-Device-Id":       XDeviceId,
		"Client-Request-Id": ClientRequestId,
		"Client-Session-Id": ClientSessionId,
		"Client-Version":    ClientVersion,
		"x-kpsdk-ct":        kpsdkct,
		"x-kpsdk-cd":        kpsdkcd,
		"Origin":            "https://www.twitch.tv",
		"DNT":               "1",
		"Connection":        "keep-alive",
		"Sec-Fetch-Dest":    "empty",
		"Sec-Fetch-Mode":    "cors",
		"Sec-Fetch-Site":    "same-site",
		"Content-Length":    "0",
	}

	req, err := http.NewRequest("POST", "https://gql.twitch.tv/integrity", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	var proxyURL *url.URL

	if shared.Config.Proxy == "your_proxy" || shared.Config.Proxy == "" {
		// Warning, if you are not using proxy, the requests can be blocked
		fmt.Println("!! There is no proxy configuration found. The requests are going to be handled without any proxy !!")
	} else {
		var err error
		proxyURL, err = url.Parse(shared.Config.Proxy)
		if err != nil {
			return nil, err
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	var cookiesReturn string

	for _, cookieData := range resp.Header["Set-Cookie"] {
		p1 := strings.Split(cookieData, ";")[0]
		cookiesReturn += cookiesReturn + p1 + "; "
	}

	tokenReturn := shared.Token{}
	err = json.Unmarshal(body, &tokenReturn)
	if err != nil {
		return nil, err
	}

	publicIntegrityData := &shared.PublicIntegrityData{
		Cookies: cookiesReturn,
		Token:   tokenReturn.Token,
	}

	return publicIntegrityData, nil
}

func getTwitchCookies() (map[string]string, error) {
	cookiesMap := make(map[string]string)
	httpClient := &http.Client{}
	var proxyURL *url.URL

	if shared.Config.Proxy == "your_proxy" || shared.Config.Proxy == "" {
		// Warning, if you are not using proxy, the requests can be blocked
		fmt.Println("!! There is no proxy configuration found. The requests are going to be handled without any proxy !!")
	} else {
		var err error
		proxyURL, err = url.Parse(shared.Config.Proxy)
		if err != nil {
			return nil, err
		}

		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	req, err := http.NewRequest("GET", "https://twitch.tv", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "current_useragent")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	for _, cookieData := range resp.Header["Set-Cookie"] {
		cookie := strings.Split(cookieData, ";")[0]
		cookiesMap[strings.Split(cookie, "=")[0]] = strings.Split(cookie, "=")[1]
	}

	return cookiesMap, nil
}
