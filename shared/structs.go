package shared

import "github.com/ox-y/GoGmailnator"

type RandomRegisterData struct {
	Username              string   `json:"username"`
	Password              string   `json:"password"`
	Birthday              Birthday `json:"birthday"`
	Email                 string   `json:"email"`
	EmailVerificationCode *string  `json:"email_verification_code"`
	ClientID              string   `json:"client_id"`
	IntegrityToken        string   `json:"integrity_token"`
	IsPasswordGuide       string   `json:"is_password_guide"`
}

type Birthday struct {
	Day      int  `json:"day"`
	Month    int  `json:"month"`
	Year     int  `json:"year"`
	IsOver18 bool `json:"is_over_18"`
}

type CreateKasadaTask struct {
	ApiKey string `json:"api_key"`
	Task   Task   `json:"task"`
}
type GetTaskResult struct {
	ApiKey string `json:"api_key"`
	TaskId string `json:"taskId"`
}

type Task struct {
	Type   string `json:"type"`
	Pjs    string `json:"pjs"`
	CdOnly bool   `json:"cdOnly"`
}

type CreateTaskResponse struct {
	ErrorCode        int    `json:"error_code"`
	ErrorDescription string `json:"error_description"`
	TaskId           string `json:"taskId"`
}

type ResultTaskResponse struct {
	ErrorId  int               `json:"errorId"`
	Solution map[string]string `json:"solution"`
	Status   string            `json:"status"`
}

type IntegrityInfo struct {
	Token   string
	Cookies map[string]string
}

type Token struct {
	Token string `json:"token"`
}

type MailnatorData struct {
	Session GoGmailnator.Session
	Email   string
}

type AccountRegisterResponse struct {
	AccessToken  string `json:"access_token"`
	RedirectPath string `json:"redirect_path"`
	UserId       string `json:"userID"`
}

type PublicIntegrityData struct {
	Token   string
	Cookies string
}

type Error struct {
	Code     interface{} `json:"code"`
	Typename string      `json:"__typename"`
}

type Request struct {
	Status   string `json:"status"`
	Typename string `json:"__typename"`
}

type ValidateVerificationCode struct {
	Error    Error   `json:"error"`
	Request  Request `json:"request"`
	Typename string  `json:"__typename"`
}

type Extensions struct {
	DurationMilliseconds int    `json:"durationMilliseconds"`
	OperationName        string `json:"operationName"`
	RequestID            string `json:"requestID"`
}

type Data struct {
	ValidateVerificationCode ValidateVerificationCode `json:"validateVerificationCode"`
}

type VerificationCodeResponse struct {
	Data       Data       `json:"data"`
	Extensions Extensions `json:"extensions"`
}

// Follow Bot

type TwitchOperationQuery struct { // This is a struct for Twitch generics requests. Various requests will use this struct.
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
	Extensions    map[string]interface{} `json:"extensions"`
}

type ErrorResponse struct {
	Breached  bool     `json:"breached"`
	Error     string   `json:"error"`
	Errors    []string `json:"errors"`
	ErrorCode int      `json:"error_code"`
}
