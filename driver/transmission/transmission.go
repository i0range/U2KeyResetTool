package transmission

import (
	"U2KeyResetTool/u2"
	"fmt"
	"github.com/hekmon/transmissionrpc"
	"strings"
)

type TransmissionDriver struct {
}

func (t TransmissionDriver) NewClient(config *u2.Config) (u2.DriverClient, error) {
	c := &DriverClient{
		config: config,
	}
	c.init()
	return c, nil
}

type DriverClient struct {
	config *u2.Config
	client *transmissionrpc.Client
}

func (c *DriverClient) init() {
	c.client = makeClient(c.config.Host, c.config.Port, c.config.Secure, c.config.User, c.config.Pass)
}

func makeClient(host string, port uint16, https bool, user string, pass string) *transmissionrpc.Client {
	conf := transmissionrpc.AdvancedConfig{
		HTTPS: https,
		Port:  port,
	}

	client, err := transmissionrpc.New(host, user, pass, &conf)
	if err != nil {
		panic(err)
	}
	return client
}

func (c *DriverClient) Check() (bool, error) {
	ok, serverVersion, minimumVersion, err := c.client.RPCVersion()
	if err != nil {
		fmt.Println("Connected to transmission server!")
		fmt.Printf("Server version %d|Server minium version %d\n", serverVersion, minimumVersion)
	}
	return ok, err
}

func (c *DriverClient) GetTorrentList(tracker string) *[]u2.Torrent {
	torrents, err := c.client.TorrentGetAll()
	if err != nil {
		fmt.Println("Error while getting torrents list!")
		panic(err)
	}

	var u2Torrents []*transmissionrpc.Torrent
	for _, torrent := range torrents {
		if len(torrent.Trackers) > 0 && strings.Contains(torrent.Trackers[0].Announce, tracker) {
			u2Torrents = append(u2Torrents, torrent)
		}
	}
	fmt.Printf("Found %d torrent(s) to process!\n", len(u2Torrents))

	var finalTorrents []u2.Torrent
	if len(u2Torrents) > 0 {
		for _, u2Torrent := range u2Torrents {
			finalTorrents = append(finalTorrents, u2.Torrent{
				Hash:    u2Torrent.HashString,
				ExtInfo: u2Torrent,
			})
		}
	}

	return &finalTorrents
}

func (c *DriverClient) EditTorrentTracker(torrent *u2.Torrent, newTracker string) (bool, error) {
	realTorrent := torrent.ExtInfo.(transmissionrpc.Torrent)
	payload := transmissionrpc.TorrentSetPayload{
		IDs:           []int64{*realTorrent.ID},
		TrackerRemove: []int64{realTorrent.Trackers[0].ID},
	}
	err := c.client.TorrentSet(&payload)
	if err != nil {
		fmt.Printf("Error while changing torrent %d %s %s\n", *realTorrent.ID, *realTorrent.HashString, *realTorrent.Name)
		fmt.Println(err)
		return false, err
	}

	payload.TrackerRemove = nil
	payload.TrackerAdd = []string{newTracker}
	err = c.client.TorrentSet(&payload)

	if err != nil {
		fmt.Printf("Error while changing torrent %d %s %s\n", *realTorrent.ID, *realTorrent.HashString, *realTorrent.Name)
		fmt.Println(err)
		return false, err
	} else {
		fmt.Printf("Change success! %d %s %s\n", *realTorrent.ID, *realTorrent.HashString, *realTorrent.Name)
		return true, nil
	}
}

func init() {
	u2.Register("transmission", &TransmissionDriver{})
}
