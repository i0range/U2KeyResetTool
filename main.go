package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hekmon/transmissionrpc"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	configFileName        = "config.json"
	processRecordFileName = "record.json"
	endpoint              = "https://u2.dmhy.org/jsonrpc_torrentkey.php?apikey="
	toTracker             = "https://daydream.dmhy.best/announce?secure="
	batchSize             = 100
)

var (
	transmissionClient *transmissionrpc.Client
	apiKey             string
	httpClient         = &http.Client{}
	silentMode         = false
)

func initClient() {
	commandConfig := parseFlag()
	if commandConfig != nil {
		silentMode = true
		apiKey = commandConfig.ApiKey
		makeClient(commandConfig.Host, commandConfig.Port, commandConfig.Secure, commandConfig.User, commandConfig.Pass)
		checkVersion()
		return
	}

	reader := bufio.NewReader(os.Stdin)

	config := readConfig()
	if config != nil {
		fmt.Println("Finding config:")
		fmt.Printf("Host: %s\nPort: %d\nHTTPS: %t\nUser: %s\nPassword %s\nAPI Key: %s\nHTTP Proxy: %s\n", config.Host, config.Port, config.Secure, config.User, config.Pass, config.ApiKey, config.Proxy)

		fmt.Print("Use this config?(y/n)")
		useConfig, _ := reader.ReadString('\n')
		useConfig = strings.TrimSpace(useConfig)

		if strings.ToLower(useConfig) == "y" || strings.ToLower(useConfig) == "yes" {
			apiKey = config.ApiKey
			setHttpProxy(config.Proxy)
			makeClient(config.Host, config.Port, config.Secure, config.User, config.Pass)
			checkVersion()
			return
		}
	}

	fmt.Print("Transmission Host [127.0.0.1]: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1"
	}
	host = extractIp(host)

	fmt.Print("Transmission Port [9091]: ")
	portString, _ := reader.ReadString('\n')
	portString = strings.TrimSpace(portString)
	if portString == "" {
		portString = "9091"
	}

	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		fmt.Println("Port invalid!")
		panic(err)
	}

	fmt.Print("Use https (y/n) [n]: ")
	useHttps, _ := reader.ReadString('\n')
	useHttps = strings.TrimSpace(useHttps)
	if useHttps == "" {
		useHttps = "n"
	}
	https := false
	if strings.ToLower(useHttps) == "y" || strings.ToLower(useHttps) == "yes" {
		https = true
	}

	fmt.Print("RPC User []: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Print("RPC Password []: ")
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)

	fmt.Print("API Key (Get From https://u2.dmhy.org/privatetorrents.php) []: ")
	key, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(key)

	fmt.Print("HTTP Proxy (May need to access U2 API, e.g. http://127.0.0.1:1080)[]: ")
	proxy, _ := reader.ReadString('\n')
	proxy = strings.TrimSpace(proxy)

	setHttpProxy(proxy)
	makeClient(host, uint16(port), https, user, pass)

	defer func() {
		if err := recover(); err != nil {
			if https {
				fmt.Printf("Please check your transmission server https://%s:%d\n", host, port)
			} else {
				fmt.Printf("Please check your transmission server http://%s:%d\n", host, port)
			}
			panic(err)
		}
	}()
	checkVersion()
	saveConfig(host, uint16(port), https, user, pass, apiKey, proxy)
}

func parseFlag() *Config {
	host := flag.String("h", "", "Transmission host")
	port := flag.Uint64("p", 0, "Transmission port")
	https := flag.Bool("s", false, "Use HTTPS")
	user := flag.String("u", "", "RPC User")
	pass := flag.String("P", "", "RPC Pass")
	key := flag.String("k", "", "U2 API Key")
	proxy := flag.String("proxy", "", "Http proxy address, i.e.: http://127.0.0.1:123")

	flag.Parse()

	setHttpProxy(*proxy)

	config := Config{
		Host:   *host,
		Port:   uint16(*port),
		Secure: *https,
		User:   *user,
		Pass:   *pass,
		ApiKey: *key,
	}

	if config.validate() {
		return &config
	}
	return nil
}

func makeClient(host string, port uint16, https bool, user string, pass string) {
	conf := transmissionrpc.AdvancedConfig{
		HTTPS: https,
		Port:  port,
	}

	client, err := transmissionrpc.New(host, user, pass, &conf)
	if err != nil {
		panic(err)
	}
	transmissionClient = client
}

func checkVersion() {
	ok, serverVersion, minimumVersion, err := transmissionClient.RPCVersion()
	if err != nil {
		fmt.Println("Error while connecting to transmission server!")
		panic(err)
	}
	if !ok {
		fmt.Println("Server too new!")
		panic("Unsupported server!")
	}
	fmt.Println("Connected to transmission server!")
	fmt.Printf("Server version %d|Server minium version %d\n", serverVersion, minimumVersion)
}

func readConfig() *Config {
	configBytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil
	}
	var config Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Println("Error while decoding saved config!")
		return nil
	}
	if config.validate() {

		return &config
	}
	return nil
}

func saveConfig(host string, port uint16, https bool, user, pass, apiKey, proxy string) {
	config := Config{
		Host:   host,
		Port:   port,
		Secure: https,
		User:   user,
		Pass:   pass,
		ApiKey: apiKey,
		Proxy:  proxy,
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		fmt.Println("Error while saving config! Json dump failed!")
		panic(err)
	}
	err = ioutil.WriteFile(configFileName, configBytes, os.FileMode(0644))
	if err != nil {
		fmt.Println("Write config failed!")
		panic(err)
	}
}

func readTorrents() []*transmissionrpc.Torrent {
	torrents, err := transmissionClient.TorrentGetAll()
	if err != nil {
		fmt.Println("Error while getting torrents list!")
		panic(err)
	}

	var u2Torrents []*transmissionrpc.Torrent
	for _, torrent := range torrents {
		if len(torrent.Trackers) > 0 && strings.Contains(torrent.Trackers[0].Announce, "dmhy") {
			u2Torrents = append(u2Torrents, torrent)
		}
	}
	fmt.Printf("Found %d torrent(s) to process!\n", len(u2Torrents))
	return u2Torrents
}

func mutateTorrentKey(torrents []*transmissionrpc.Torrent) {
	records := readRecords()
	var needProcessTorrents []*transmissionrpc.Torrent
	for _, torrent := range torrents {
		if _, ok := records[*torrent.HashString]; ok {
			continue
		}
		needProcessTorrents = append(needProcessTorrents, torrent)
	}

	fmt.Printf("Find %d torrent(s) to process!\n", len(needProcessTorrents))

	for {
		count := 0
		var requestData []U2Request
		torrentMap := make(map[int]*transmissionrpc.Torrent)
		for _, torrent := range needProcessTorrents {
			count += 1
			requestData = append(requestData, U2Request{
				JsonRpc: "2.0",
				Method:  "query",
				Params:  []string{*torrent.HashString},
				Id:      count,
			})
			torrentMap[count] = torrent

			if count == batchSize {
				doMutate(records, requestData, torrentMap)
				count = 0
				requestData = []U2Request{}
				torrentMap = make(map[int]*transmissionrpc.Torrent)
				fmt.Println("Wait 5 seconds for next batch.")
				time.Sleep(5 * time.Second)
			}
		}

		if count > 0 {
			doMutate(records, requestData, torrentMap)
		}

		break
	}
}

func doMutate(records map[string]int, data []U2Request, torrentMap map[int]*transmissionrpc.Torrent) {
	jsonRequestBytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Process u2 request failed!")
		fmt.Println(err)
		return
	}

	req, err := http.NewRequest("POST", endpoint+apiKey, bytes.NewBuffer(jsonRequestBytes))
	if err != nil {
		fmt.Println("Process u2 request failed!")
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	retryCount := 0
	for {
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Process u2 request failed!")
			fmt.Println(err)
			return
		}

		fmt.Println("response Status:", resp.Status)
		if resp.StatusCode == 200 {
			body, _ := ioutil.ReadAll(resp.Body)
			closeBody(resp)

			var secretKeyResponse []U2Response

			err := json.Unmarshal(body, &secretKeyResponse)
			if err != nil {
				fmt.Println("Error while processing u2 response!")
			} else {
				for _, response := range secretKeyResponse {
					if response.Id > 0 && response.Result != "" {
						if updateTorrent(torrentMap[response.Id], response.Result) {
							records[*(torrentMap[response.Id].HashString)] = 1
						}
					} else {
						fmt.Println("Skip torrent because of response error!")
						fmt.Printf("%d %s\n", response.Error.Code, response.Error.Message)
					}
				}
				saveRecords(records)
			}
			break
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
				keepWindow(-1)

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
}

func updateTorrent(torrent *transmissionrpc.Torrent, secretKey string) bool {
	payload := transmissionrpc.TorrentSetPayload{
		IDs:           []int64{*torrent.ID},
		TrackerRemove: []int64{torrent.Trackers[0].ID},
	}
	err := transmissionClient.TorrentSet(&payload)
	if err != nil {
		fmt.Printf("Error while changing torrent %d %s %s\n", *torrent.ID, *torrent.HashString, *torrent.Name)
		fmt.Println(err)
		return false
	}

	payload.TrackerRemove = nil
	payload.TrackerAdd = []string{toTracker + secretKey}
	err = transmissionClient.TorrentSet(&payload)

	if err != nil {
		fmt.Printf("Error while changing torrent %d %s %s\n", *torrent.ID, *torrent.HashString, *torrent.Name)
		fmt.Println(err)
		return false
	} else {
		fmt.Printf("Change success! %d %s %s\n", *torrent.ID, *torrent.HashString, *torrent.Name)
		return true
	}
}

func readRecords() map[string]int {
	records := make(map[string]int)
	recordBytes, err := ioutil.ReadFile(processRecordFileName)
	if err != nil {
		return records
	}

	err = json.Unmarshal(recordBytes, &records)
	if err != nil {
		records = make(map[string]int)
	}
	if records == nil {
		records = make(map[string]int)
	}
	return records
}

func saveRecords(records map[string]int) {
	recordsBytes, err := json.Marshal(records)
	if err != nil {
		fmt.Println("Error while saving records! Json dump failed!")
		panic(err)
	}
	err = ioutil.WriteFile(processRecordFileName, recordsBytes, os.FileMode(0644))
	if err != nil {
		fmt.Println("Write records failed!")
		panic(err)
	}
}

func extractIp(host string) string {
	host = strings.ReplaceAll(host, "http://", "")
	host = strings.ReplaceAll(host, "https://", "")
	host = strings.ReplaceAll(host, "/", "")
	return host
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

func closeBody(resp *http.Response) {
	if resp.Body != nil {
		resp.Body.Close()
	}
}

func keepWindow(code int) {
	if !silentMode {
		fmt.Println("Finished! Press enter key to exit!")
		fmt.Scanln()
	}
	os.Exit(code)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			keepWindow(-1)
		} else {
			keepWindow(0)
		}
	}()
	initClient()
	torrents := readTorrents()
	mutateTorrentKey(torrents)
}
