# U2KeyResetTool

A small tool written in Golang to replace Tracker address for Transmission v2.x, qBittorrent v4.x and Deluge v1.x.

Tested on Transmission v2.94, qBittorrent v4.2.5 and Deluge v1.3.15

## Origin Python Version：
Tr-U2@ITGR(https://gist.github.com/inertia42/f6120d118e47925095dbceb5e8e27272)  

qB-U2@杯杯杯杯具(https://gist.github.com/tongyifan/83220b417cffdd23528860ee0c518d15)

De-U2@種崎敦美(https://github.com/XSky123/dmhy_change_securekey_deluge)

## Change
1. Completely rewritten in Golang
2. Add interactive config input function

## How to use
1. Get your u2 API Key at https://u2.dmhy.org/privatetorrents.php
2. Download the latest release at https://github.com/i0range/U2KeyResetTool/releases
3. Run U2KeyResetTool

## Command Line Arguments
|Argument|Type  |Required|Usage          |
| ------ | ---- | ------ | ------------- |
|-t      |string|Required|Target program, t for Transmission, q for qBittorrent, d for Deluge (default "t")|
|-h      |string|Required|Host IP Address|
|-p      |uint  |Required|Port           |
|-s      |bool  |Optional|Use HTTPS      |
|-u      |string|Optional|Username       |
|-P      |string|Optional|Password       |
|-k      |string|Required|U2 API Key     |
|-proxy  |string|Optional|Http proxy address, e.g.: http://127.0.0.1:123|

For example, reset key for torrents on Transmission server on 192.168.1.2 port 9091 with user admin pass admin should use this command:

```./U2KeyResetTool -t t -h 192.168.1.2 -p 9091 -u admin -P admin -k __YOUR_KEY__```

If your server need https, just add `-s` flag

## How to build
1. Install Golang (Only tested on 1.15)
2. Clone code
3. Run `go build`

### Optional
1. Install goreleaser
2. Run `goreleaser --rm-dist --snapshot` to build for all platform

## Changelog
https://github.com/i0range/U2KeyResetTool/releases