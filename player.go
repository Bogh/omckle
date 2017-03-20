package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/anacrolix/missinggo"
	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"

	"github.com/anacrolix/torrent"
)

type PlayerState uint8

type Player struct {
	Torrent *torrent.Torrent
	File    *torrent.File
	Reader  io.ReadSeeker

	cancel context.CancelFunc
	cmd    *exec.Cmd
	out    io.ReadCloser
	in     io.WriteCloser

	state PlayerState
}

const (
	readaheadSize        = 1024 * 1024      // 1MB
	playerStartThreshold = 10 * 1024 * 1024 // 10MB

	StateUnavailable PlayerState = iota
	StatePlaying
	StatePaused
	StateStopped
)

var (
	ActivePlayer *Player

	actions = map[string]string{
		"pause": "pause",
	}
)

// Play stream in mplayer
func Play(t *torrent.Torrent) error {
	if ActivePlayer != nil {
		if err := ActivePlayer.Close(); err != nil {
			log.Println("Error closing Active Player: ", err)
			return err
		}
	}

	f := getMovieFile(t)
	f.Download()

	// Prioritize
	f.PrioritizeRegion(0, playerStartThreshold)

	tReader := t.NewReader()
	tReader.SetReadahead(readaheadSize)
	r := missinggo.NewSectionReadSeeker(tReader, f.Offset(), f.Length())
	ActivePlayer = &Player{
		Torrent: t,
		File:    f,
		Reader:  r,
	}

	// Start the player once we read at least 10MB
	go ActivePlayer.watchAndRun()
	// go ActivePlayer.outputStats()
	return nil
}

func (p *Player) outputStats() {
	for {
		select {
		case <-p.Torrent.Closed():
			log.Println("Torrent ", p.Torrent.Name(), " has been closed.")
			return
		case <-time.After(5 * time.Second):
			info := p.Torrent.Info()
			log.Printf(`
                    Torrent Stats:
                        Name: %s
                        Bytes Completed: %s
                        Bytes Missing: %s
                        Length: %s
                `,
				p.Torrent.Name(),
				humanize.Bytes(uint64(p.Torrent.BytesCompleted())),
				humanize.Bytes(uint64(p.Torrent.BytesMissing())),
				humanize.Bytes(uint64(info.TotalLength())))
		}

	}

}

func (p *Player) Close() error {
	log.Println("Closing player...")

	// stop command if running
	if p.cancel != nil {
		p.cancel()
	}

	if closer, ok := p.Reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	p.File.Cancel()
	p.Torrent.Drop()
	return nil
}

// Run Mplayer for stream
func (p *Player) Run() error {
	log.Println("Running mplayer")

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	// TODO Bogdan: Get dynamic stream link
	cmd := exec.CommandContext(ctx,
		"mplayer", "-quiet", "-slave", "http://localhost:8080/stream")
	cmd.Stdout = os.Stdout

	if in, err := cmd.StdinPipe(); err == nil {
		p.in = in
	} else {
		log.Println("Cannot obtain player stdin: ", err)
		return err
	}

	go func() {
		if err := p.cmd.Run(); err != nil {
			log.Println("Running player command error: ", err)
		}
	}()

	p.cmd = cmd
	return nil
}

// watchAndRun watches the torrent stats and runs the player once `playerStartThreshold`
// number of byes have been downloaded
func (p *Player) watchAndRun() {
	for {
		log.Println("Checking download status. Waiting for threshold to download")
		if p.Torrent.BytesCompleted() > playerStartThreshold {
			p.Run()
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func (p *Player) action(c *gin.Context, action string) error {
	if p.in == nil {
		return fmt.Errorf("No `in` pipe for player.")
	}

	if f, ok := actions[action]; !ok {
		return fmt.Errorf("Unknown action: %s.", action)
	} else {
		action = f
	}

	action += "\n"

	if _, err := io.WriteString(p.in, action); err != nil {
		log.Printf("Error executing action %s: %s\n", action, err)
		return err
	}
	log.Println("Action executed: ", action)

	return nil
}

func PlayerAPIAction(c *gin.Context) {
	if ActivePlayer == nil {
		c.Error(fmt.Errorf("No active player. Must upload a torrent first."))
		return
	}
	action := c.Param("action")

	if err := ActivePlayer.action(c, action); err != nil {
		c.Error(err)
		return
	}
}
