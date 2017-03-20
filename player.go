package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/anacrolix/missinggo"
	humanize "github.com/dustin/go-humanize"

	"github.com/anacrolix/torrent"
)

const (
	readaheadSize        = 1024 * 1024      // 1MB
	playerStartThreshold = 10 * 1024 * 1024 // 10MB
)

var ActivePlayer *Player

type Player struct {
	Torrent *torrent.Torrent
	File    *torrent.File
	Reader  io.ReadSeeker

	cmd *exec.Cmd

	cancel context.CancelFunc
}

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
	go ActivePlayer.outputStats()
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
	p.cmd = exec.CommandContext(
		ctx, "mplayer", "http://localhost:8080/stream")
	p.cmd.Stdout = os.Stdout
	go func() {
		if err := p.cmd.Run(); err != nil {
			log.Println("Running player command error: ", err)
		}
	}()

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
