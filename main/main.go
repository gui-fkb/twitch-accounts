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
	"strconv"
	"strings"
	"sync"
	"time"

	"twitch-accounts/shared"

	"github.com/goombaio/namegenerator"
	"github.com/ox-y/GoGmailnator"
	"github.com/sethvargo/go-password/password"
)

var outputFile string = "./results/accounts.txt"
var userPassFile string = "./results/userpass.txt"
var tokensFile string = "./results/tokens.txt"

func main() {
	fmt.Println("twitch-accounts by xBadApple -  https://github.com/xBadApple")
	//shared.FastEmailTest() // Uncomment this line if you want to test the trash email in a fast way, dont forget to enable breakpoints inside the function

	if shared.Config.CapSolverKey == "your_captcha_key" {
		log.Fatal("It looks like your captcha solver API token isn't configured yet. Change it in the shared.Config.go file and run again.")
	}

	fmt.Println("Starting account creation...")
	fmt.Println("How many accounts do you want to create?")
	var quantity int
	_, err := fmt.Scanln(&quantity)
	if err != nil {
		fmt.Println("Invalid input. Please enter a number.")
		return
	}

	var wg sync.WaitGroup
	sem := make(chan bool, 4) // Limit to 4 concurrent goroutines

	for i := 0; i < quantity; i++ {
		wg.Add(1)
		go func(i int) {
			sem <- true // Will block if there is already 5 goroutines running
			defer func() {
				<-sem // Release the slot
				wg.Done()
			}()

			fmt.Printf("Creating account %d of %d\n", i+1, quantity)
			createNewAccount()
		}(i)
	}

	wg.Wait()
	close(sem)

	fmt.Println("Finished accounts creation.")
}

func createNewAccount() {
	randomUsername := getRandomUsername() + "_" + shared.GenerateRandomID(3)

	trashMailSession, err := getTrashMailSession()
	if err != nil {
		fmt.Println(err, "\n account creation exited")
		return
	}

	randomEmail := trashMailSession.Email

	registerPostData := generateRandomRegisterData(randomUsername, randomEmail)

	fmt.Println("Getting twitch cookies.")
	cookies, err := getTwitchCookies()
	if err != nil {
		fmt.Println(err, "\n account creation exited")
		return
	}

	fmt.Println("Getting kasada code")
	taskResponse, err := kasadaResolver()
	if err != nil {
		fmt.Println(err, "\n account creation exited")
		return
	}

	fmt.Println("Getting local integrity token")
	err = getIntegrityOption(taskResponse)
	if err != nil {
		fmt.Println(err, "\n account creation exited")
		return
	}

	integrityData, err := integrityGetToken(taskResponse, cookies)
	fmt.Printf("IntegrityToken: %v", integrityData.Token[:48]+"+"+strconv.FormatInt(int64(len(integrityData.Token)-48), 10)+"... \n")
	if err != nil {
		fmt.Println(err, "\n unable to register token - account creation exited")
		return
	}

	fmt.Println("Creating account...")
	registerPostData.IntegrityToken = integrityData.Token
	registerData, err := registerFinal(cookies, registerPostData, taskResponse.Solution["user-agent"])
	if err != nil {
		fmt.Println(err, "\n error creating account - account creation exited")
		return
	}

	userId := registerData.UserId
	accessToken := registerData.AccessToken

	fmt.Println("Account created!")
	fmt.Println("UserID:", userId, "AccessToken:", accessToken)

	fmt.Println("Waiting email verification ...")
	time.Sleep(time.Second * 8) // Sleep for 8 seconds because twitch verification email can have some delay
	verifyCode, err := getVerificationCode(trashMailSession)
	if err != nil {
		fmt.Println(err, "\n error getting verification code - account creation exited")
		return
	}

	verifyEmailResponse := &shared.VerificationCodeResponse{}

	for i := 0; i < 4; i++ {
		fmt.Println("Getting Kasada Code attempt", i+1, " of ", 4)
		kasada2, err := kasadaResolver()
		if err != nil {
			fmt.Println(err, "\n error getting kasada code - account creation exited")
			continue
		}

		clientSessionId := shared.GenerateRandomID(16)
		xDeviceId := cookies["unique_id"]
		clientVersion := "3040e141-5964-4d72-b67d-e73c1cf355b5"
		clientRequestId := shared.GenerateRandomID(32)

		fmt.Println("Getting public integrity token...")
		publicIntegrityData, err := publicIntegrityGetToken(xDeviceId, clientRequestId, clientSessionId, clientVersion, kasada2.Solution["x-kpsdk-ct"], kasada2.Solution["x-kpsdk-cd"], accessToken, kasada2.Solution["user-agent"])
		fmt.Printf("PublicIntegrityToken: %v", publicIntegrityData.Token[:48]+"+"+strconv.FormatInt(int64(len(publicIntegrityData.Token)-48), 10)+"... \n")
		if err != nil {
			fmt.Println(err, "\n error getting public integrity token - account creation exited")
			continue
		}

		fmt.Println("Verifying account email...")
		verifyEmailResponse, err = verifyEmail(xDeviceId, clientVersion, clientSessionId, accessToken, publicIntegrityData.Token, verifyCode, userId, trashMailSession.Email, kasada2.Solution["user-agent"])
		if err != nil {
			fmt.Println(err, "\n error verifying account email - account creation exited")
			continue
		}

		if verifyEmailResponse == nil {
			fmt.Println(err, "\n email verification failed - account creation exited")
			continue
		}

		if verifyEmailResponse.Data.ValidateVerificationCode.Request.Status == "VERIFIED" {
			break
		}
	}

	if verifyEmailResponse.Data.ValidateVerificationCode.Request.Status == "VERIFIED" {
		err := saveAccountData(registerPostData, userId, accessToken)
		if err != nil {
			fmt.Println(err, "\n error saving account data - account creation exited")
			return
		}

		fmt.Println("Account verified and saved!")
	} else {
		fmt.Println("Account WAS NOT NOT VERIFIED")
		jsonBytes, _ := json.Marshal(verifyEmailResponse)

		fmt.Println("Verify email response: " + string(jsonBytes))

		return
	}

	fmt.Println("Account is ready!")
}

func getRandomUsername() string {
	nameGenerator := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

	name := strings.Replace(nameGenerator.Generate(), "-", "", -1)
	return name
}

func getEmail(username string) string { // This function is not being used right now, but it can be useful in the future
	return fmt.Sprintf("%s@%s", username, shared.Config.EmailDomain)
}

func generateRandomRegisterData(uname string, email string) shared.RandomRegisterData {
	return shared.RandomRegisterData{
		Username:       uname,
		Password:       getRandomPassword(),
		Birthday:       generateRandomBirthday(),
		Email:          email,
		ClientID:       shared.Config.TwitchClientID,
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

func generateRandomBirthday() shared.Birthday {
	return shared.Birthday{
		Day:   rand.Intn(25) + 1,
		Month: rand.Intn(12) + 1,
		Year:  rand.Intn(30) + 1970,
	}
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

func getIntegrityOption(taskResponse *shared.ResultTaskResponse) error {
	client := &http.Client{}
	var proxyURL *url.URL

	if shared.Config.Proxy == "your_proxy" || shared.Config.Proxy == "" {
		// Warning, if you are not using proxy, the requests can be blocked
		fmt.Println("!! There is no proxy configuration found. The requests are going to be handled without any proxy !!")
	} else {
		var err error
		proxyURL, err = url.Parse(shared.Config.Proxy)
		if err != nil {
			return err
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	req, err := http.NewRequest("OPTIONS", "https://passport.twitch.tv/integrity", nil)
	if err != nil {
		return err
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

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func integrityGetToken(taskResponse *shared.ResultTaskResponse, cookies map[string]string) (*shared.Token, error) {
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

	req, err := http.NewRequest("POST", "https://passport.twitch.tv/integrity", nil)
	if err != nil {
		return nil, err
	}

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
		return nil, err
	}
	defer resp.Body.Close()

	for _, cookieData := range resp.Header["Set-Cookie"] {
		cookie := strings.Split(cookieData, ";")[0]
		cookies[strings.Split(cookie, "=")[0]] = strings.Split(cookie, "=")[1]
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	token := &shared.Token{}
	err = json.Unmarshal(body, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func registerFinal(cookies map[string]string, postParams shared.RandomRegisterData, userAgent string) (*shared.AccountRegisterResponse, error) {
	var cookiesString string
	for key, value := range cookies {
		cookiesString += key + "=" + value + "; "
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

	jsonBody, err := json.Marshal(postParams)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://passport.twitch.tv/protected_register", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

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

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		registerResponse := &shared.AccountRegisterResponse{}
		err = json.Unmarshal(body, registerResponse)
		if err != nil {
			return nil, err
		}

		return registerResponse, nil
	} else {
		return nil, errors.New(string(body))
	}
}

func getTrashMailSession() (*shared.MailnatorData, error) {
	var sess GoGmailnator.Session

	var proxy *string
	if shared.Config.Proxy == "your_proxy" || shared.Config.Proxy == "" {
		// Warning, if you are not using proxy, the requests can be blocked
		fmt.Println("!! There is no proxy configuration found. The requests are going to be handled without any proxy !!")
		proxy = nil
	} else {
		tempProxy := strings.Replace(strings.Replace(shared.Config.Proxy, "https://", "", -1), "http://", "", -1) // Remove https:// from the proxy, because the GoGmailnator package is hardcoded to use http
		proxy = &tempProxy
	}

	// session will expire after a few hours
	err := sess.Init(proxy)
	if err != nil {
		return nil, err
	}

	// calling sess.GenerateEmailAddress or sess.RetrieveMail with a dead session will cause an error
	isAlive, err := sess.IsAlive()
	if err != nil {
		return nil, err
	}

	if isAlive {
		fmt.Println("Session is alive.")
	} else {
		fmt.Println("Session is dead.")
		return nil, fmt.Errorf("session is not alive")
	}

	emailAddress, err := sess.GenerateEmailAddress()
	if err != nil {
		return nil, err
	}

	fmt.Println("Email address is " + emailAddress + ".")

	mailData := &shared.MailnatorData{
		Session: sess,
		Email:   emailAddress,
	}

	return mailData, nil
}

func getVerificationCode(mailData *shared.MailnatorData) (string, error) {
	emails, err := mailData.Session.RetrieveMail(mailData.Email)
	if err != nil {
		return "", err
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

func verifyEmail(XDeviceId, ClientVersion, ClientSessionId, accessToken, ClientIntegrity, code, userId, email, current_useragent string) (*shared.VerificationCodeResponse, error) {
	query := `{"operationName":"ValidateVerificationCode","variables":{"input":{"code":"` + code + `","key":"` + userId + `","address":"` + email + `"}},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"05eba55c37ee4eff4dae260850dd6703d99cfde8b8ec99bc97a67e584ae9ec31"}}}`

	requestBody := bytes.NewBufferString(query)

	headers := map[string]string{
		"User-Agent":       current_useragent,
		"Accept":           "application/json",
		"Accept-Language":  "en-US",
		"Accept-Encoding":  "identity",
		"Referer":          "https://www.twitch.tv/",
		"Client-Id":        shared.Config.TwitchClientID,
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

	verificationResponse := &shared.VerificationCodeResponse{}
	if err := json.Unmarshal(body, &verificationResponse); err != nil {
		return nil, err
	}

	return verificationResponse, nil
}

func saveAccountData(r shared.RandomRegisterData, userId string, accesToken string) error {
	// 1. Save Full data

	// Check if the file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		// If the file doesn't exist, create an empty file
		if err := os.WriteFile(outputFile, []byte(""), 0644); err != nil {
			return err
		}
	}
	dataAll := r.Username + " " + r.Password + " " + r.Email + " " + userId + " " + accesToken + " \n"
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	// Write data to the file
	if _, err := file.Write([]byte(dataAll)); err != nil {
		return err
	}

	// 2. Save only the username and password
	if _, err := os.Stat(userPassFile); os.IsNotExist(err) {
		if err := os.WriteFile(userPassFile, []byte(""), 0644); err != nil {
			return err
		}
	}
	userPass := r.Username + " " + r.Password + " \n"
	file, err = os.OpenFile(userPassFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write([]byte(userPass)); err != nil {
		return err
	}

	// 3. Save only the tokens (oauth acces token)
	if _, err := os.Stat(tokensFile); os.IsNotExist(err) {
		if err := os.WriteFile(tokensFile, []byte(""), 0644); err != nil {
			return err
		}
	}
	token := accesToken + " \n"
	file, err = os.OpenFile(tokensFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write([]byte(token)); err != nil {
		return err
	}

	return nil
}
