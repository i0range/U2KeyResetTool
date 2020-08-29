package qBittorrent

import (
	"U2KeyResetTool/u2"
	"fmt"
	qBittorrent "github.com/i0range/go-qbittorrent"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type TorrentInfo struct {
	Hash    string
	Name    string
	Tracker string
}

type Driver struct {
}

type DriverClient struct {
	config *u2.Config
	client *qBittorrent.Client
}

func (c *DriverClient) Check() (bool, error) {
	version, err := c.client.Application.GetAPIVersion()
	if err != nil {
		return false, err
	}
	fmt.Printf("Current qBittorrent API version %s\n", version)
	return true, nil
}

func (c *DriverClient) GetTorrentList(tracker string) *[]u2.Torrent {
	torrents, err := c.client.Torrent.GetList(nil)
	if err != nil {
		fmt.Println("Error while getting torrent list from qBittorrent!")
		panic(err)
	}

	var u2Torrents []TorrentInfo
	for _, torrent := range torrents {
		trackers, err := c.client.Torrent.GetTrackers(torrent.Hash)
		if err != nil {
			fmt.Printf("Getting tracker of torrent %s %s failed!\n", torrent.Hash, torrent.Name)
			fmt.Println(err)
			continue
		}
		if len(trackers) > 0 {
			for _, torrentTracker := range trackers {
				if strings.Contains(torrentTracker.URL, tracker) {
					u2Torrents = append(u2Torrents, TorrentInfo{
						Hash:    torrent.Hash,
						Name:    torrent.Name,
						Tracker: torrentTracker.URL,
					})
					break
				}
			}
		}
	}
	fmt.Printf("Found %d torrent(s) to process!\n", len(u2Torrents))

	var finalTorrents []u2.Torrent
	if len(u2Torrents) > 0 {
		for _, u2Torrent := range u2Torrents {
			finalTorrents = append(finalTorrents, u2.Torrent{
				Hash:    u2Torrent.Hash,
				ExtInfo: u2Torrent,
			})
		}
	}

	return &finalTorrents
}

func (c *DriverClient) EditTorrentTracker(torrent *u2.Torrent, tracker string) (bool, error) {
	realTorrent := torrent.ExtInfo.(TorrentInfo)

	err := c.client.Torrent.EditTrackers(realTorrent.Hash, realTorrent.Tracker, tracker)
	if err != nil {
		fmt.Printf("Error while changing torrent %s %s\n", realTorrent.Hash, realTorrent.Name)
		fmt.Println(err)
		return false, err
	} else {
		fmt.Printf("Change success! %s %s\n", realTorrent.Hash, realTorrent.Name)
		return true, nil
	}
}

func (q Driver) NewClient(config *u2.Config) (u2.DriverClient, error) {
	return &DriverClient{
		config: config,
		client: makeClient(config),
	}, nil
}

func makeClient(config *u2.Config) *qBittorrent.Client {
	var baseUrl string
	if config.Secure {
		baseUrl += "https://"
	} else {
		baseUrl += "http://"
	}
	baseUrl += config.Host + ":" + strconv.Itoa(int(config.Port))
	client := qBittorrent.NewClient(baseUrl, log.New())
	if config.User != "" {
		err := client.Login(config.User, config.Pass)
		if err != nil {
			fmt.Printf("Error while connecting to qBittorrent %s\n", baseUrl)
			panic(err)
		}
	}
	return client
}

func init() {
	u2.Register("qBittorrent", &Driver{})
}
