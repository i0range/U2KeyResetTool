package u2

import "strings"

type Torrent struct {
	Hash    string
	ExtInfo interface{}
}

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
	Target string
	Host   string
	Port   uint16
	Secure bool
	User   string
	Pass   string
	ApiKey string
	Proxy  string
}

func (c *Config) Validate() bool {
	if c.Target == "" {
		c.Target = "transmission"
	}
	if c.Host == "" {
		return false
	}
	if c.Port == 0 {
		return false
	}
	if c.ApiKey == "" {
		return false
	}
	c.ApiKey = strings.ReplaceAll(c.ApiKey, "https://", "")
	c.ApiKey = strings.ReplaceAll(c.ApiKey, "u2.dmhy.org", "")
	c.ApiKey = strings.ReplaceAll(c.ApiKey, "/", "")
	c.ApiKey = strings.ReplaceAll(c.ApiKey, "jsonrpc_torrentkey.php", "")
	c.ApiKey = strings.ReplaceAll(c.ApiKey, "apikey=", "")
	c.ApiKey = strings.ReplaceAll(c.ApiKey, "?", "")
	return true
}
