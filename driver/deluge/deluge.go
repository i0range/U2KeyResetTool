package deluge

import (
	"U2KeyResetTool/u2"
	"fmt"
	deluge "github.com/gdm85/go-libdeluge"
	"strings"
	"time"
)

type Driver struct {
}

func (d Driver) NewClient(config *u2.Config) (u2.DriverClient, error) {
	return &DriverClient{
		config: config,
		client: makeClient(config),
	}, nil
}

func makeClient(config *u2.Config) *deluge.Client {
	client := deluge.NewV1(deluge.Settings{
		Hostname:             config.Host,
		Port:                 uint(config.Port),
		Login:                config.User,
		Password:             config.Pass,
		ReadWriteTimeout:     60 * time.Second,
		DebugServerResponses: false,
	})
	return client
}

type DriverClient struct {
	config *u2.Config
	client *deluge.Client
}

func (c *DriverClient) Check() (bool, error) {
	err := c.client.Connect()
	if err != nil {
		fmt.Printf("Error while connecting to Deluge %s %d as user %s!\n", c.config.Host, c.config.Port, c.config.User)
		return false, err
	}
	version, err := c.client.DaemonVersion()
	if err != nil {
		fmt.Println("Error while reading Deluge version!")
		return false, err
	}
	fmt.Printf("Current deluge version %s\n", version)
	return true, nil
}

func (c *DriverClient) GetTorrentList(tracker string) *[]u2.Torrent {
	torrents, err := c.client.TorrentsStatus("", []string{})
	if err != nil {
		fmt.Println("Error while getting torrents info!")
		fmt.Println(err)
		panic(err)
	}

	var finalTorrents []u2.Torrent
	for hash, torrent := range torrents {
		if strings.Contains(torrent.TrackerHost, tracker) {
			finalTorrents = append(finalTorrents, u2.Torrent{
				Hash:    hash,
				ExtInfo: *torrent,
			})
		}
	}

	fmt.Printf("Found %d torrent(s) from Deluge!\n", len(finalTorrents))
	return &finalTorrents
}

func (c *DriverClient) EditTorrentTracker(torrent *u2.Torrent, newTracker string) (bool, error) {
	realTorrent := torrent.ExtInfo.(deluge.TorrentStatus)
	err := c.client.SetTorrentTracker(torrent.Hash, newTracker)
	if err != nil {
		fmt.Printf("Error while changing torrent %s %s\n", torrent.Hash, realTorrent.Name)
		fmt.Println(err)
		return false, err
	} else {
		fmt.Printf("Change success! %s %s\n", torrent.Hash, realTorrent.Name)
		return true, nil
	}
}

func init() {
	u2.Register("deluge", &Driver{})
}
