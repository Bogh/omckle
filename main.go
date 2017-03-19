package main

import (
	"log"
	"net/http"
	"time"

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
			case "file":
				file, header, err := c.Request.FormFile("upload")
				if err != nil {
					log.Println("Error reading request file: ", err)
					ctx["Error"] = "Cannot read uploaded file. Invalid file uploaded."
					break
				}

				filename := header.Filename
				log.Println("Uploaded file: ", filename)

				if err = AddFromReader(file); err != nil {
					ctx["Error"] = "Cannot add torrent from uploaded file."
					break
				}
			}
			c.HTML(http.StatusOK, "index.gohtml", ctx)
		})

	stream := router.Group("/stream")
	{
		stream.GET("/", func(c *gin.Context) {
			if ActivePlayer == nil {
				c.String(http.StatusNotFound, "No video found.")
				return
			}

			http.ServeContent(c.Writer, c.Request, ActivePlayer.File.Path(), time.Time{}, ActivePlayer.Reader)
		})
	}
	router.Run(":8080")
}
