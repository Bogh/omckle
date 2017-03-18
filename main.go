package main

import (
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
