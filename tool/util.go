package tool

import (
	"fmt"
	"os"
	"strings"
)

func ExtractIp(host string) string {
	host = strings.ReplaceAll(host, "http://", "")
	host = strings.ReplaceAll(host, "https://", "")
	host = strings.ReplaceAll(host, "/", "")
	return host
}

func ParseTarget(target string) string {
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

func KeepWindow(code int) {
	if !silentMode {
		fmt.Println("Finished! Press enter key to exit!")
		_, _ = fmt.Scanln()
	}
	os.Exit(code)
}
