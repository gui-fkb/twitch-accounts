package main

type Config struct {
	CapSolverKey   string
	Proxy          string
	UserAgent      string
	EmailDomain    string
	TwitchClientID string
	SalamonderKey  string
}

type RandomRegisterData struct {
	Username       string
	Password       string
	Birthday       Birthday
	Email          string
	ClientID       string
	IntegrityToken string
}

type Birthday struct {
	Day   int
	Month int
	Year  int
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
