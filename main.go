package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/anacrolix/missinggo"
	"github.com/anacrolix/torrent"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Name = "Omckle"
	app.Usage = "torrent streaming"
	app.ArgsUsage = "<file>"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "Enable debug logging",
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		if c.NArg() == 0 {
			log.Error("Torrent file not specified.")
			return fmt.Errorf("Torrent file is required.")
		}

		path := c.Args().First()
		log.WithField("file", path).Debug("Fetching torrent file")

		client, err := torrent.NewClient(&torrent.Config{Debug: true})
		if err != nil {
			log.WithError(err).Error("Error creating torrent client")
			return err
		}

		t, err := client.AddTorrentFromFile(path)
		if err != nil {
			log.WithError(err).Error("Error adding torrent file")
			return err
		}

		// Wait for info
		<-t.GotInfo()
		t.DownloadAll()
		mi := t.Info()
		log.WithFields(log.Fields{
			"name":   mi.Name,
			"length": mi.Length,
			"source": mi.Source,
			"files":  len(mi.Files),
		}).Info("Torrent metainfo")

		movieFile := getMovieFile(t)
		log.WithField("moviefile", movieFile.Path()).Info("Movie File")

		r := t.NewReader()
		defer r.Close()

		rs := missinggo.NewSectionReadSeeker(r, movieFile.Offset(), movieFile.Length())
		p := make([]byte, 0, 1024)
		n, err := rs.Read(p)
		if err != nil {
			log.WithError(err).Error("Error reading torrent file")
			return err
		}
		log.WithField("lenght", n).Info("Read number of bytes")

		for {

			fmt.Println("Bytes completed: ", t.BytesCompleted())
			time.Sleep(time.Second)
		}
		return nil
	}

	app.Run(os.Args)
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
