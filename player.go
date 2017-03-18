package main

import (
	"io"
	"log"

	"github.com/anacrolix/missinggo"

	"github.com/anacrolix/torrent"
)

var ActivePlayer *Player

type Player struct {
	Torrent *torrent.Torrent
	File    *torrent.File
	Reader  io.ReadSeeker
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
	r := missinggo.NewSectionReadSeeker(t.NewReader(), f.Offset(), f.Length())
	ActivePlayer = &Player{t, f, r}
	return nil
}

func (p *Player) Close() error {
	log.Println("Closing player...")
	if closer, ok := p.Reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return err
		}
	}

	p.File.Cancel()
	return nil
}
