package main

import (
	_ "U2KeyResetTool/driver/qBittorrent"
	_ "U2KeyResetTool/driver/transmission"
	"U2KeyResetTool/u2"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	configFileName        = "config.json"
	processRecordFileName = "record.json"
	toTracker             = "https://daydream.dmhy.best/announce?secure="
	batchSize             = 100
)

var (
	apiKey     string
	silentMode = false
	client     *u2.Client
)

func initClient() {
	commandConfig := parseFlag()
	if commandConfig != nil {
		silentMode = true
		apiKey = commandConfig.ApiKey

		u2Client, err := u2.NewClient(commandConfig)
		if err != nil {
			fmt.Println("Error while creating client!")
			fmt.Println(err)
			panic(err)
		}
		client = u2Client
		checkVersion()
		return
	}

	reader := bufio.NewReader(os.Stdin)

	config := readConfig()
	if config != nil {
		fmt.Println("Finding config:")
		fmt.Printf("Target: %s\nHost: %s\nPort: %d\nHTTPS: %t\nUser: %s\nPassword: %s\nAPI Key: %s\nHTTP Proxy: %s\n", config.Target, config.Host, config.Port, config.Secure, config.User, config.Pass, config.ApiKey, config.Proxy)

		fmt.Print("Use this config?(y/n)")
		useConfig, _ := reader.ReadString('\n')
		useConfig = strings.TrimSpace(useConfig)

		if strings.ToLower(useConfig) == "y" || strings.ToLower(useConfig) == "yes" {
			apiKey = config.ApiKey
			makeU2Client(config)
			checkVersion()
			return
		}
	}

	fmt.Println("t for Transmission, q for qBittorrent, d for Deluge")
	fmt.Print("Target program (t/q/d) [t]:")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)
	target = parseTarget(target)

	fmt.Print("Host [127.0.0.1]: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1"
	}
	host = extractIp(host)

	fmt.Print("Port [9091]: ")
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

	fmt.Print("User []: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Print("Password []: ")
	pass, _ := reader.ReadString('\n')
	pass = strings.TrimSpace(pass)

	fmt.Print("API Key (Get From https://u2.dmhy.org/privatetorrents.php) []: ")
	key, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(key)

	fmt.Print("HTTP Proxy (May need to access U2 API, e.g. http://127.0.0.1:1080)[]: ")
	proxy, _ := reader.ReadString('\n')
	proxy = strings.TrimSpace(proxy)

	u2Config := u2.Config{
		Target: target,
		Host:   host,
		Port:   uint16(port),
		Secure: https,
		User:   user,
		Pass:   pass,
		ApiKey: apiKey,
		Proxy:  proxy,
	}

	u2Config.Validate()

	makeU2Client(&u2Config)

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
	saveConfig(&u2Config)
}

func makeU2Client(config *u2.Config) {
	u2Client, err := u2.NewClient(config)
	if err != nil {
		panic(err)
	}
	client = u2Client
}

func parseFlag() *u2.Config {
	target := flag.String("t", "t", "Target program, t for Transmission, q for qBittorrent, d for Deluge")
	host := flag.String("h", "", "Host")
	port := flag.Uint64("p", 0, "Port")
	https := flag.Bool("s", false, "Use HTTPS")
	user := flag.String("u", "", "User")
	pass := flag.String("P", "", "Pass")
	key := flag.String("k", "", "U2 API Key")
	proxy := flag.String("proxy", "", "Http proxy address, i.e.: http://127.0.0.1:123")

	flag.Parse()

	config := u2.Config{
		Target: parseTarget(*target),
		Host:   *host,
		Port:   uint16(*port),
		Secure: *https,
		User:   *user,
		Pass:   *pass,
		ApiKey: *key,
		Proxy:  *proxy,
	}

	if config.Validate() {
		return &config
	}
	return nil
}

func checkVersion() {
	ok, err := client.Check()
	if err != nil {
		fmt.Println("Error while connecting to transmission server!")
		panic(err)
	}
	if !ok {
		fmt.Println("Server too new!")
		panic("Unsupported server!")
	}
}

func readConfig() *u2.Config {
	configBytes, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil
	}
	var config u2.Config
	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Println("Error while decoding saved config!")
		return nil
	}
	if config.Validate() {

		return &config
	}
	return nil
}

func saveConfig(config *u2.Config) {
	configBytes, err := json.Marshal(*config)
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

func readTorrents() *[]u2.Torrent {
	return client.GetTorrentList("dmhy")
}

func mutateTorrentKey(torrents *[]u2.Torrent) {
	records := readRecords()
	var needProcessTorrents []u2.Torrent
	for _, torrent := range *torrents {
		if _, ok := records[torrent.Hash]; ok {
			continue
		}
		needProcessTorrents = append(needProcessTorrents, torrent)
	}

	fmt.Printf("Find %d torrent(s) to process!\n", len(needProcessTorrents))

	for {
		count := 0
		var requestData []u2.U2Request
		torrentMap := make(map[int]u2.Torrent)
		for _, torrent := range needProcessTorrents {
			count += 1
			requestData = append(requestData, u2.U2Request{
				JsonRpc: "2.0",
				Method:  "query",
				Params:  []string{torrent.Hash},
				Id:      count,
			})
			torrentMap[count] = torrent

			if count == batchSize {
				doMutate(records, &requestData, torrentMap)
				count = 0
				requestData = []u2.U2Request{}
				torrentMap = make(map[int]u2.Torrent)
				fmt.Println("Wait 5 seconds for next batch.")
				time.Sleep(5 * time.Second)
			}
		}

		if count > 0 {
			doMutate(records, &requestData, torrentMap)
		}

		break
	}
}

func doMutate(records map[string]int, data *[]u2.U2Request, torrentMap map[int]u2.Torrent) {
	secretKeyResponse, err := client.GetNewKey(data)
	if err != nil {
		fmt.Println("Error while getting new key from u2!")
		fmt.Println(err)
		keepWindow(-1)
	}
	for _, response := range *secretKeyResponse {
		if response.Id > 0 && response.Result != "" {
			if updateTorrent(torrentMap[response.Id], response.Result) {
				records[(torrentMap[response.Id].Hash)] = 1
			}
		} else {
			fmt.Println("Skip torrent because of response error!")
			fmt.Printf("%d %s\n", response.Error.Code, response.Error.Message)
		}
	}
	saveRecords(records)
}

func updateTorrent(torrent u2.Torrent, secretKey string) bool {
	return client.EditTorrentTracker(&torrent, toTracker+secretKey)
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

func parseTarget(target string) string {
	target = strings.ToLower(target)
	if strings.HasPrefix(target, "t") {
		return "transmission"
	} else if strings.HasPrefix(target, "q") {
		return "qBittorrent"
	} else if strings.HasPrefix(target, "d") {
		return "deluge"
	}
	return ""
}

func keepWindow(code int) {
	if !silentMode {
		fmt.Println("Finished! Press enter key to exit!")
		_, _ = fmt.Scanln()
	}
	os.Exit(code)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error while changing key!")
			fmt.Println(err)
			keepWindow(-1)
		} else {
			keepWindow(0)
		}
	}()
	initClient()
	torrents := readTorrents()
	mutateTorrentKey(torrents)
}
