package main

import (
	"io"
	"log"
	"os"
	"path"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

var Client *torrent.Client

func InitClient() error {
	dataDir := path.Join(os.TempDir(), "omckle")
	var err error
	Client, err = torrent.NewClient(&torrent.Config{
		DataDir: dataDir,
		Debug:   true,
	})
	if err != nil {
		log.Println("Error creating torrent client: ", err)
		return err
	}

	log.Println("Torrent data: ", len(Client.Torrents()))
	log.Println("Torrent data dir: ", dataDir)
	return nil
}

func AddMagnetLink(magnet string) error {
	t, err := Client.AddMagnet(magnet)
	if err != nil {
		log.Println("Cannot load torrent from magnet file: ", err)
		return err
	}

	return initTorrent(t)
}

func AddFromReader(r io.Reader) error {
	mi, err := metainfo.Load(r)
	if err != nil {
		log.Println("Error reading meta info from Reader: ", err)
		return err
	}
	t, err := Client.AddTorrent(mi)
	if err != nil {
		log.Println("Cannot load torrent from reader: ", err)
		return err
	}
	return initTorrent(t)
}

func initTorrent(t *torrent.Torrent) error {
	<-t.GotInfo()
	log.Printf("Torrent added: %s\n", t.Name())

	return Play(t)
}

func getMovieFile(t *torrent.Torrent) *torrent.File {
	var f torrent.File
	var s int64

	for _, file := range t.Files() {
		if s < file.Length() {
			s = file.Length()
			f = file
		}
	}

	return &f
}
