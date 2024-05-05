package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/sethvargo/go-password/password"
)

func main() {
	fmt.Println("twitch-accounts by xBadApple -  https://github.com/xBadApple")

	if config.CapSolverKey == "your_captcha_key" {
		log.Fatal("It looks like your captcha solver API token isn't configured yet. Change it in the config.go file and run again.")
	}

	createNewAccount()

}

func createNewAccount() {
	randomUsername := getRandomUsername() + "_" + generateRandomID(3)
	randomEmail := getEmail(randomUsername)

	registerPostData := generateRandomRegisterData(randomUsername, randomEmail)

	fmt.Println("Getting twitch cookies.")
	cookies := getTwitchCookies()

	fmt.Println("Getting kasada code")
	taskResponse := kasadaResolver()

	fmt.Println("Getting local integrity token") // Add proxy later into integrity
	getIntegrityOption(taskResponse)

	integrityData := integrityGetToken(taskResponse)
	if integrityData.Token == "" {
		log.Fatal("Unable to get register token!")
	}

	fmt.Println("Creating account...")
	registerPostData.IntegrityToken = integrityData.Token

	fmt.Printf("%+v", registerPostData)
	fmt.Printf("%+v", cookies)
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
		index := rand.Intn(len(charset))
		bytes[i] = charset[index]
	}
	return string(bytes)
}

func getEmail(username string) string {
	return fmt.Sprintf("%s@%s", username, config.EmailDomain)
}

func generateRandomRegisterData(uname string, email string) RandomRegisterData {
	return RandomRegisterData{
		Username:       uname,
		Password:       getRandomPassword(),
		Birthday:       generateRandomBirthday(),
		Email:          email,
		ClientID:       config.TwitchClientID,
		IntegrityToken: "",
	}
}

func getRandomPassword() string {
	res, err := password.Generate(32, 1, 1, false, false)
	if err != nil {
		log.Fatal(err)
	}

	return res
}

func generateRandomBirthday() Birthday {
	return Birthday{
		Day:   rand.Intn(30) + 1,
		Month: rand.Intn(12) + 1,
		Year:  rand.Intn(30) + 1970,
	}
}

func getTwitchCookies() map[string]string {
	cookiesMap := make(map[string]string)
	httpClient := &http.Client{}
	var proxyURL *url.URL

	if config.Proxy == "your_proxy" {
		fmt.Println("!! There is no proxy configuration found. The requests are going to be handled without any proxy !!")
	} else {
		var err error
		proxyURL, err = url.Parse(config.Proxy)
		if err != nil {
			log.Fatal("Error parsing proxy URL:", err)
		}

		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	req, err := http.NewRequest("GET", "https://twitch.tv", nil)
	if err != nil {
		log.Fatal("Error creating the request:", err)
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
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	for _, cookieData := range resp.Header["Set-Cookie"] {
		cookie := strings.Split(cookieData, ";")[0]
		cookiesMap[strings.Split(cookie, "=")[0]] = strings.Split(cookie, "=")[1]
	}

	return cookiesMap
}

func kasadaResolver() ResultTaskResponse {
	taskResponse := createKasadaTask()
	time.Sleep(time.Second * 5)
	taskResult := getTaskResult(taskResponse.TaskId)

	fmt.Println(taskResult)

	return taskResult
}

func createKasadaTask() CreateTaskResponse {
	requestBody := CreateKasadaTask{
		ApiKey: config.SalamonderKey,
		Task: Task{
			Type:   "KasadaCaptchaSolver",
			Pjs:    "https://k.twitchcdn.net/149e9513-01fa-4fb0-aad4-566afd725d1b/2d206a39-8ed7-437e-a3be-862e0f06eea3/p.js",
			CdOnly: false,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post("https://salamoonder.com/api/createTask", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(body))

	taskResp := CreateTaskResponse{}
	json.Unmarshal(body, &taskResp)

	return taskResp
}

func getTaskResult(taskId string) ResultTaskResponse {
	task := GetTaskResult{TaskId: taskId}

	jsonBody, _ := json.Marshal(task)

	resp, _ := http.Post("https://salamoonder.com/api/getTaskResult", "application/json", bytes.NewBuffer(jsonBody))
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	taskResponse := ResultTaskResponse{}

	json.Unmarshal(body, &taskResponse)

	return taskResponse
}

func getIntegrityOption(taskResponse ResultTaskResponse) {
	client := &http.Client{}

	req, err := http.NewRequest("OPTIONS", "https://passport.twitch.tv/integrity", nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	req.Header.Set("User-Agent", taskResponse.Solution["user-agent"])
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "x-kpsdk-cd,x-kpsdk-ct")
	req.Header.Set("Referer", "https://www.twitch.tv/")
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error sending request:", err)
	}

	defer resp.Body.Close()

	// Print the response status code
	fmt.Println("Response Status:", resp.Status)
}

func integrityGetToken(taskResponse ResultTaskResponse) IntegrityInfo {
	client := &http.Client{}

	req, _ := http.NewRequest("POST", "https://passport.twitch.tv/integrity", nil)

	req.Header.Set("User-Agent", taskResponse.Solution["user-agent"])
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://www.twitch.tv/")
	req.Header.Set("x-kpsdk-ct", taskResponse.Solution["x-kpsdk-ct"])
	req.Header.Set("x-kpsdk-cd", taskResponse.Solution["x-kpsdk-cd"])
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Content-Length", "0")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	integrityInfo := IntegrityInfo{}

	for _, cookieData := range resp.Header["Set-Cookie"] {
		cookie := strings.Split(cookieData, ";")[0]
		integrityInfo.Cookies[strings.Split(cookie, "=")[0]] = strings.Split(cookie, "=")[1]
	}

	body, _ := io.ReadAll(resp.Body)

	integrityInfo.Token = string(body)

	return integrityInfo
}
