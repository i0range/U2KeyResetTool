# U2KeyResetTool

A small tool written in Golang to replace Tracker address for Transmission v2.x, qBittorrent v4.x and Deluge v1.x.

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

## How to build
1. Install Golang (Only tested on 1.15)
2. Clone code
3. Run `go build`

### Optional
1. Install goreleaser
2. Run `goreleaser --rm-dist --snapshot` to build for all platform

## Changelog
https://github.com/i0range/U2KeyResetTool/releases