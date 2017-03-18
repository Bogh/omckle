package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dustin/go-humanize"

	"github.com/anacrolix/torrent"
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
		return err
	}

	<-t.GotInfo()
	log.Printf("Torrent added: %s\n", t.Name())

	Play(t)

	// Output torrent stats
	go func() {
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
	}()
	return nil
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
