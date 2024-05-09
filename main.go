package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/goombaio/namegenerator"
	"github.com/ox-y/GoGmailnator"
	"github.com/sethvargo/go-password/password"
)

var outputFile string = "./results/accounts.txt"

func main() {
	fmt.Println("twitch-accounts by xBadApple -  https://github.com/xBadApple")
	//fastEmailTest() // Uncomment this line if you want to test the trash email in a fast way, dont forget to enable breakpoints inside the function

	if config.CapSolverKey == "your_captcha_key" {
		log.Fatal("It looks like your captcha solver API token isn't configured yet. Change it in the config.go file and run again.")
	}

	createNewAccount()
}

func createNewAccount() {
	randomUsername := getRandomUsername() + "_" + generateRandomID(3)

	trashMailSession := getTrashMailSession()
	randomEmail := trashMailSession.Email

	registerPostData := generateRandomRegisterData(randomUsername, randomEmail)

	fmt.Println("Getting twitch cookies.")
	cookies := getTwitchCookies()

	fmt.Println("Getting kasada code")

	taskResponse := kasadaResolver()
	fmt.Println("Getting local integrity token") // Add proxy later into integrity
	getIntegrityOption(taskResponse)

	integrityData := integrityGetToken(taskResponse, cookies)
	if integrityData.Token == "" {
		log.Fatal("Unable to get register token!")
	}

	fmt.Println("Creating account...")
	registerPostData.IntegrityToken = integrityData.Token
	registerData, err := registerFinal(cookies, registerPostData, taskResponse.Solution["user-agent"])
	if err != nil {
		log.Fatal(err)
	}

	userId := registerData.UserId
	accessToken := registerData.AccessToken

	fmt.Println("Account created!")
	fmt.Println("UserID:", userId, "AccessToken:", accessToken)

	fmt.Println("Waiting email verification ...")
	time.Sleep(time.Second * 10) // Sleep for 10 seconds because twitch verification email can have some delay
	verifyCode, _ := getVerificationCode(trashMailSession)

	fmt.Println("Getting Kasada Code")
	kasada2 := kasadaResolver()

	clientSessionId := generateRandomID(16)
	xDeviceId := cookies["unique_id"]
	clientVersion := "3040e141-5964-4d72-b67d-e73c1cf355b5"
	clientRequestId := generateRandomID(32)

	fmt.Println("Getting public integrity token...")
	publicIntegrityData, err := publicIntegrityGetToken(xDeviceId, clientRequestId, clientSessionId, clientVersion, kasada2.Solution["x-kpsdk-ct"], kasada2.Solution["x-kpsdk-cd"], accessToken, kasada2.Solution["user-agent"])
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Verifying account email...")
	verifyEmailResponse, err := verifyEmail(xDeviceId, clientVersion, clientSessionId, accessToken, publicIntegrityData.Token, verifyCode, userId, trashMailSession.Email, kasada2.Solution["user-agent"])
	if err != nil {
		log.Fatal(err)
	}

	if verifyEmailResponse == nil {
		log.Fatal("Email verification failed!")
	}

	if verifyEmailResponse.Data.ValidateVerificationCode.Request.Status == "VERIFIED" {
		saveAccountData(registerPostData, userId, accessToken)
		fmt.Println("Account verified and saved!")
	}

	fmt.Println("Account is ready!")
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
	return fmt.Sprintf("%s@%s", username, config.EmailDomain) // Unused right now
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
	time.Sleep(time.Second * 2)
	taskResult := getTaskResult(taskResponse.TaskId)

	fmt.Println(taskResult)

	return taskResult
}

func createKasadaTask() CreateTaskResponse {
	requestBody := CreateKasadaTask{
		ApiKey: config.CapSolverKey,
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

func integrityGetToken(taskResponse ResultTaskResponse, cookies map[string]string) Token {
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

	for _, cookieData := range resp.Header["Set-Cookie"] {
		cookie := strings.Split(cookieData, ";")[0]
		cookies[strings.Split(cookie, "=")[0]] = strings.Split(cookie, "=")[1]
	}

	body, _ := io.ReadAll(resp.Body)

	token := Token{}
	json.Unmarshal(body, &token)

	return token
}

func registerFinal(cookies map[string]string, postParams RandomRegisterData, userAgent string) (*AccountRegisterResponse, error) {
	var cookiesString string
	for key, value := range cookies {
		cookiesString += key + "=" + value + "; "
	}

	client := &http.Client{}

	jsonBody, _ := json.Marshal(postParams)

	req, _ := http.NewRequest("POST", "https://passport.twitch.tv/protected_register", bytes.NewBuffer(jsonBody))

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://www.twitch.tv/")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Origin", "https://www.twitch.tv")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", cookiesString)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")

	resp, _ := client.Do(req)

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		registerResponse := &AccountRegisterResponse{}
		json.Unmarshal(body, registerResponse)

		return registerResponse, nil
	} else {
		return nil, errors.New(string(body))
	}

}

func getTrashMailSession() (*MailnatorData, err) {
	var sess GoGmailnator.Session

	// session will expire after a few hours
	err := sess.Init(nil)
	if err != nil {
		nil, err
	}

	// calling sess.GenerateEmailAddress or sess.RetrieveMail with a dead session will cause an error
	isAlive, err := sess.IsAlive()
	if err != nil {
		nil, err
	}

	if isAlive {
		fmt.Println("Session is alive.")
	} else {
		fmt.Println("Session is dead.")
		return nil, err
	}

	emailAddress, err := sess.GenerateEmailAddress()
	if err != nil {
		nil, err
	}

	fmt.Println("Email address is " + emailAddress + ".")

	mailData := &MailnatorData{
		Session: sess,
		Email:   emailAddress,
	}

	return mailData, nil
}

func getVerificationCode(mailData *MailnatorData) (string, error) {
	emails, err := mailData.Session.RetrieveMail(mailData.Email)
	if err != nil {
		panic(err)
	}

	var verificationCode string
	for _, email := range emails {
		if strings.Contains(email.Subject, "Twitch") {
			split := strings.Split(email.Subject, "–")[0]
			verificationCode = strings.TrimSpace(split)
			break
		}
	}

	if verificationCode == "" {
		return "", errors.New("there is no twitch email")
	}

	fmt.Println("Verification code:", verificationCode)

	return verificationCode, nil
}

func publicIntegrityGetToken(XDeviceId, ClientRequestId, ClientSessionId, ClientVersion, kpsdkct, kpsdkcd, accesstoken, current_useragent string) (publicIntegrity *PublicIntegrityData, err error) {
	requestBody := []byte("{}")

	headers := map[string]string{
		"User-Agent":        current_useragent,
		"Accept":            "application/json",
		"Accept-Language":   "en-US",
		"Accept-Encoding":   "identity",
		"Authorization":     "OAuth " + accesstoken,
		"Referer":           "https://www.twitch.tv/",
		"Client-Id":         config.TwitchClientID,
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

	tokenReturn := Token{}
	json.Unmarshal(body, &tokenReturn)

	publicIntegrityData := &PublicIntegrityData{
		Cookies: cookiesReturn,
		Token:   tokenReturn.Token,
	}

	return publicIntegrityData, nil
}

func verifyEmail(XDeviceId, ClientVersion, ClientSessionId, accessToken, ClientIntegrity, code, userId, email, current_useragent string) (*VerificationCodeResponse, error) {
	query := `{"operationName":"ValidateVerificationCode","variables":{"input":{"code":"` + code + `","key":"` + userId + `","address":"` + email + `"}},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"05eba55c37ee4eff4dae260850dd6703d99cfde8b8ec99bc97a67e584ae9ec31"}}}`

	requestBody := bytes.NewBufferString(query)

	headers := map[string]string{
		"User-Agent":       current_useragent,
		"Accept":           "application/json",
		"Accept-Language":  "en-US",
		"Accept-Encoding":  "identity",
		"Referer":          "https://www.twitch.tv/",
		"Client-Id":        config.TwitchClientID,
		"X-Device-Id":      XDeviceId,
		"Client-Version":   ClientVersion,
		"Client-Session":   ClientSessionId,
		"Authorization":    "OAuth " + accessToken,
		"Client-Integrity": ClientIntegrity,
		"Content-Type":     "text/plain;charset=UTF-8",
		"Origin":           "https://www.twitch.tv",
		"DNT":              "1",
		"Connection":       "keep-alive",
		"Sec-Fetch-Dest":   "empty",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Site":   "same-site",
	}

	req, err := http.NewRequest("POST", "https://gql.twitch.tv/gql#origin=twilight", requestBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d", resp.StatusCode)
	}

	verificationResponse := &VerificationCodeResponse{}
	if err := json.Unmarshal(body, &verificationResponse); err != nil {
		return nil, fmt.Errorf("error parsing response JSON: %v", err)
	}

	return verificationResponse, nil
}

func saveAccountData(r RandomRegisterData, userId string, accesToken string) {
	// Check if the file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		// If the file doesn't exist, create an empty file
		if err := os.WriteFile(outputFile, []byte(""), 0644); err != nil {
			panic(err)
		}
	}

	dataAll := r.Username + " " + r.Password + " " + r.Email + " " + userId + " " + accesToken + "\n"

	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write data to the file
	if _, err := file.Write([]byte(dataAll)); err != nil {
		panic(err)
	}
}
