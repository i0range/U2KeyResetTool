package main

type U2Request struct {
	JsonRpc string   `json:"jsonrpc"`
	Method  string   `json:"method"`
	Params  []string `json:"params"`
	Id      int      `json:"id"`
}

type U2Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type U2Response struct {
	Id     int     `json:"id,omitempty"`
	Result string  `json:"result,omitempty"`
	Error  U2Error `json:"error,omitempty"`
}

type Config struct {
	Host   string
	Port   uint16
	Secure bool
	User   string
	Pass   string
	ApiKey string
	Proxy  string
}

func (c Config) validate() bool {
	if c.Host == "" {
		return false
	}
	if c.Port == 0 {
		return false
	}
	if c.ApiKey == "" {
		return false
	}
	return true
}
