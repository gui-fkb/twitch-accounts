package main

type Config struct {
	CapSolverKey string
	Proxy        string
	UserAgent    string
	EmailDomain  string
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
