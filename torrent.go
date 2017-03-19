package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

var Client *torrent.Client

func init() {
	dataDir, err := ioutil.TempDir(os.TempDir(), "omckle")
	if err != nil {
		log.Fatalln("Cannot create temporary directory: ", err)
	}
	Client, err = torrent.NewClient(&torrent.Config{
		DataDir: dataDir,
		Debug:   true,
	})
	if err != nil {
		log.Fatal("Error creating torrent client: ", err)
	}
	log.Println("Torrent data dir: ", dataDir)

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

func outputStats(t *torrent.Torrent) {
	for {
		select {
		case <-t.Closed():
			log.Println("Torrent ", t.Name(), " has been closed.")
			return
		case <-time.After(5 * time.Second):
			info := t.Info()
			log.Printf(`
                    Torrent Stats:
                        Name: %s
                        Bytes Completed: %s
                        Bytes Missing: %s
                        Length: %s
                `,
				t.Name(),
				humanize.Bytes(uint64(t.BytesCompleted())),
				humanize.Bytes(uint64(t.BytesMissing())),
				humanize.Bytes(uint64(info.TotalLength())))
		}

	}

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
