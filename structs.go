package main

import "github.com/ox-y/GoGmailnator"

type Config struct {
	CapSolverKey   string
	Proxy          string
	UserAgent      string
	EmailDomain    string
	TwitchClientID string
}

type RandomRegisterData struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	Birthday       Birthday `json:"birthday"`
	Email          string   `json:"email"`
	ClientID       string   `json:"client_id"`
	IntegrityToken string   `json:"integrity_token"`
}

type Birthday struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type CreateKasadaTask struct {
	ApiKey string `json:"api_key"`
	Task   Task   `json:"task"`
}
type GetTaskResult struct {
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
