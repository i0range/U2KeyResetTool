package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/i0range/U2KeyResetTool/driver/deluge"
	_ "github.com/i0range/U2KeyResetTool/driver/qBittorrent"
	_ "github.com/i0range/U2KeyResetTool/driver/transmission"
	"github.com/i0range/U2KeyResetTool/u2"
	"io/ioutil"
	"os"
	"time"
)

const (
	processRecordFileName = "record.json"
	toTracker             = "https://daydream.dmhy.best/announce?secure="
	batchSize             = 100
)

var (
	silentMode = false
	client     *u2.Client
)

func ProcessTorrent() {
	torrents := readTorrents()
	mutateTorrentKey(torrents)
}

func InitClient(config *u2.Config) {
	config.Validate()
	makeU2Client(config)

	defer func() {
		if err := recover(); err != nil {
			if config.Secure {
				fmt.Printf("Please check your %s server https://%s:%d\n", config.Target, config.Host, config.Port)
			} else {
				fmt.Printf("Please check your %s server http://%s:%d\n", config.Target, config.Host, config.Port)
			}
			panic(err)
		}
	}()
	checkVersion()
	saveConfig(config)
}

func TurnOnSilentMode() {
	silentMode = true
}
func makeU2Client(config *u2.Config) {
	u2Client, err := u2.NewClient(config)
	if err != nil {
		fmt.Println("Error while creating client!")
		fmt.Println(err)
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

	fmt.Printf("Found %d torrent(s) to process!\n", len(needProcessTorrents))

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
		panic(err)
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
	InitClient(initConfig())
	ProcessTorrent()
}
