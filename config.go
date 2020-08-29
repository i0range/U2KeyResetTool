package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/i0range/U2KeyResetTool/tool"
	"github.com/i0range/U2KeyResetTool/u2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	configFileName = "config.json"
)

func initConfig() *u2.Config {
	commandConfig := parseFlag()
	if commandConfig != nil {
		tool.TurnOnSilentMode()
		return commandConfig
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
			return config
		}
	}

	fmt.Println("t for Transmission, q for qBittorrent, d for Deluge")
	fmt.Print("Target program (t/q/d) [t]:")
	target, _ := reader.ReadString('\n')
	target = strings.TrimSpace(target)
	target = tool.ParseTarget(target)

	fmt.Print("Host [127.0.0.1]: ")
	host, _ := reader.ReadString('\n')
	host = strings.TrimSpace(host)
	if host == "" {
		host = "127.0.0.1"
	}
	host = tool.ExtractIp(host)

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
	apiKey := strings.TrimSpace(key)

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

	return &u2Config
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
		Target: tool.ParseTarget(*target),
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
