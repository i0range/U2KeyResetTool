package u2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var (
	driversMu  sync.RWMutex
	drivers    = make(map[string]Driver)
	endpoint   = "https://u2.dmhy.org/jsonrpc_torrentkey.php?apikey="
	httpClient = &http.Client{}
)

type Driver interface {
	NewClient(*Config) (DriverClient, error)
}

type DriverClient interface {
	Check() (bool, error)

	GetTorrentList(tracker string) *[]Torrent

	EditTorrentTracker(torrent *Torrent, newTracker string) (bool, error)
}

type Client struct {
	config     *Config
	realClient *DriverClient
}

func (c *Client) GetNewKey(data *[]U2Request) (*[]U2Response, error) {
	jsonRequestBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Process u2 request failed!")
		fmt.Println(err)
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint+c.config.ApiKey, bytes.NewBuffer(jsonRequestBytes))
	if err != nil {
		fmt.Println("Process u2 request failed!")
		fmt.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	retryCount := 0
	for {
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Process u2 request failed!")
			fmt.Println(err)
			return nil, err
		}

		fmt.Println("response Status:", resp.Status)
		if resp.StatusCode == 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			closeBody(resp)

			var secretKeyResponse []U2Response

			err := json.Unmarshal(body, &secretKeyResponse)
			if err != nil {
				fmt.Println("Error while processing u2 response!")
				return nil, err
			} else {
				return &secretKeyResponse, nil
			}
		} else {
			if resp.StatusCode == 503 {
				retryAfter := resp.Header.Get("Retry-After")
				waitSecond := 5
				if retryAfter != "" {
					retryAfterInt, err := strconv.Atoi(retryAfter)
					if err != nil {
						fmt.Println("Convert retry after failed! Use default wait time!")
					} else {
						waitSecond += retryAfterInt
					}
				}
				fmt.Printf("Rate limit! Waiting %d seconds!\n", waitSecond)
				time.Sleep(time.Duration(waitSecond) * time.Second)
			} else if resp.StatusCode == 403 {
				fmt.Println("Wrong API key! Please note: API Key IS NOT passkey!")
				if resp.Body != nil {
					body, _ := ioutil.ReadAll(resp.Body)
					fmt.Println(string(body))
				}
				return nil, fmt.Errorf("Wrong API key!")
			} else {
				fmt.Println("Unrecognized error! Retry after 5 seconds!")
				if resp.Body != nil {
					body, _ := ioutil.ReadAll(resp.Body)
					fmt.Println(string(body))
				}
				time.Sleep(5 * time.Second)
			}
			retryCount++
		}

		closeBody(resp)

		if retryCount > 5 {
			fmt.Println("Too many retried! Please check your network!")
			break
		}
	}
	return nil, fmt.Errorf("Failed after %d retries!", retryCount)
}

func (c *Client) Check() (bool, error) {
	return (*c.realClient).Check()
}

func (c *Client) GetTorrentList(tracker string) *[]Torrent {
	return (*c.realClient).GetTorrentList(tracker)
}

func (c *Client) EditTorrentTracker(torrent *Torrent, newTracker string) bool {
	ok, err := (*c.realClient).EditTorrentTracker(torrent, newTracker)
	if err != nil {
		fmt.Println("Error while edit torrent %s", torrent.Hash)
	}
	return ok
}

func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("u2: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("u2: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func NewClient(config *Config) (*Client, error) {
	driversMu.RLock()
	driverI, ok := drivers[config.Target]
	driversMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("u2: unknown driver %q (forgotten import?)", config.Target)
	}

	driverClient, err := driverI.NewClient(config)
	if err != nil {
		return nil, err
	}

	return newClient(config, &driverClient)
}

func newClient(config *Config, realClient *DriverClient) (*Client, error) {
	setHttpProxy(config.Proxy)
	return &Client{
		config:     config,
		realClient: realClient,
	}, nil
}

func closeBody(resp *http.Response) {
	if resp.Body != nil {
		resp.Body.Close()
	}
}

func setHttpProxy(proxy string) {
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			fmt.Println("Invalid proxy config")
			fmt.Println(err)
		} else {
			httpClient = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
			fmt.Printf("Using proxy %s for U2 request!\n", proxy)
		}
	}
}
