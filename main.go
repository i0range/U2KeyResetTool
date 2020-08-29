package main

import (
	"fmt"
	_ "github.com/i0range/U2KeyResetTool/driver/deluge"
	_ "github.com/i0range/U2KeyResetTool/driver/qBittorrent"
	_ "github.com/i0range/U2KeyResetTool/driver/transmission"
	"github.com/i0range/U2KeyResetTool/tool"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Error while changing key!")
			fmt.Println(err)
			tool.KeepWindow(-1)
		} else {
			tool.KeepWindow(0)
		}
	}()
	config := initConfig()
	tool.InitClient(config)
	saveConfig(config)
	tool.ProcessTorrent()
}
