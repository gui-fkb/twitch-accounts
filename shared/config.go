package shared

type Configuration struct {
	CapSolverKey   string
	Proxy          string
	UserAgent      string
	EmailDomain    string
	TwitchClientID string
}

// Config variable stores your configuration
var Config = Configuration{
	CapSolverKey:   "your_api_key",
	Proxy:          "your_proxy",
	UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0",
	TwitchClientID: "kimne78kx3ncx6brgo4mv6wki5h1ko",
}
