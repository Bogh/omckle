package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	router.
		GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.gohtml", nil)
		}).
		POST("/", func(c *gin.Context) {
			ctx := gin.H{}
			switch c.DefaultPostForm("format", "magnet") {
			case "magnet":
				magnetLink := c.PostForm("magnet")
				if magnetLink == "" {
					ctx["Error"] = "Magnet url is required."
				}
				err := AddMagnetLink(magnetLink)
				if err == nil {
					ctx["Success"] = true
				} else {
					ctx["Error"] = err.Error()
				}
			}
			c.HTML(http.StatusOK, "index.gohtml", ctx)
		})

	router.Run(":8080")
	// app := cli.NewApp()

	// app.Name = "Omckle"
	// app.Usage = "torrent streaming"
	// app.ArgsUsage = "<file>"
	// app.Flags = []cli.Flag{
	// 	cli.BoolFlag{
	// 		Name:  "debug, d",
	// 		Usage: "Enable debug logging",
	// 	},
	// }
	// app.Action = func(c *cli.Context) error {
	// 	client, err := torrent.NewClient(&torrent.Config{
	// 		DataDir:    os.TempDir(),
	// 		Debug:      true,
	// 		Seed:       true,
	// 		NoUpload:   false,
	// 		DisableTCP: false,
	// 		ListenAddr: ":50007",
	// 	})
	// 	if err != nil {
	// 		log.WithError(err).Error("Error creating torrent client")
	// 		return err
	// 	}

	// 	t, err := client.AddMagnet("magnet:?x.t=urn:btih:bc37ea8ae2d0dfec5bd9d6800b3e0af9399e4deb&dn=The.Big.Bang.Theory.S10E18.HDTV.x264-LOL&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Fzer0day.ch%3A1337&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Fpublic.popcorn-tracker.org%3A6969")
	// 	if err != nil {
	// 		log.WithError(err).Error("Error adding magnet torrent")
	// 		return err
	// 	}

	// 	go func() {
	// 		// Wait for info
	// 		<-t.GotInfo()
	// 		t.DownloadAll()

	// 		mi := t.Info()
	// 		log.WithFields(log.Fields{
	// 			"name":   mi.Name,
	// 			"length": mi.Length,
	// 			"source": mi.Source,
	// 			"files":  len(mi.Files),
	// 		}).Info("Torrent metainfo")

	// 		movieFile := getMovieFile(t)
	// 		log.WithField("moviefile", movieFile.Path()).Info("Movie File")

	// 		r := t.NewReader()
	// 		defer r.Close()

	// 		rs := missinggo.NewSectionReadSeeker(r, movieFile.Offset(), movieFile.Length())
	// 		p := make([]byte, 1024)
	// 		n, err := rs.Read(p)
	// 		if err != nil {
	// 			log.WithError(err).Error("Error reading torrent file")
	// 			return
	// 		}
	// 		log.WithField("lenght", n).Info("Read number of bytes")
	// 	}()

	// 	for {
	// 		select {
	// 		case <-t.Closed():
	// 			fmt.Println("Torrent is closed.")
	// 		default:
	// 			fmt.Println("Torrent stats: ", t.Stats())
	// 			time.Sleep(time.Second)
	// 		}
	// 	}

	// 	return nil
	// }

	// app.Run(os.Args)
}
