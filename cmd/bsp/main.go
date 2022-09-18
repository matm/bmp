package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/matm/bsp/pkg/mpd"
	"github.com/matm/bsp/pkg/types"
)

func secondsToHuman(secs int) string {
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

type bookmark struct {
	start, end string
}

func main() {
	mp := mpd.NewClient()
	defer mp.Close()
	r := bufio.NewReader(os.Stdin)

	quit := false
	// Bookmark start.
	bms := make([]bookmark, 0)
	bOpen := false

	for !quit {
		fmt.Printf("> ")
		ch, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
		}
		switch string(ch) {
		case "q", "quit", "exit":
			quit = true
			fmt.Println("Bye!")
		case "[":
			// Bookmark start.
			st, err := mp.Status()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			if st.State != "play" {
				fmt.Println("Please starting playing the song first")
				continue
			}
			if bOpen {
				fmt.Println("Missing closing bookmark, please use ']' first")
				continue
			}
			bOpen = true
			start := secondsToHuman(int(st.Elapsed))
			bms = append(bms, bookmark{start: start})
			fmt.Println(start)
		case "]":
			// Bookmark end.
			st, err := mp.Status()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			if st.State != "play" {
				fmt.Println("Please starting playing the song first")
				continue
			}
			if !bOpen {
				fmt.Println("Missing opening bookmark, please use '[' first")
				continue
			}
			bOpen = false
			end := secondsToHuman(int(st.Elapsed))
			bm := &bms[len(bms)-1]
			bm.end = end
			fmt.Printf("%s-%s\n", bm.start, bm.end)
		case "i":
			// Current song info.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			st, err := mp.Status()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			fmt.Printf("[%s] %s: %s\n", st.State, s.Artist, s.Title)
			fmt.Printf("%s/%s\n", secondsToHuman(int(st.Elapsed)), secondsToHuman(int(st.Duration)))
		case "f":
			// Forward seek +10s.
			err := mp.SeekOffset(10)
			if err != nil {
				log.Print(err)
				continue
			}
		case "b":
			// Backward seek -10s.
			st, err := mp.Status()
			if err != nil {
				log.Print(err)
				continue
			}
			// Seek to absolute time. Relative backward seeking not working as expected, whereas
			// forward seeking works well.
			err = mp.SeekTo(int(st.Elapsed) - 10)
			if err != nil {
				log.Print(err)
				continue
			}
		case "t":
			// Toogle play/pause.
			err := mp.Toggle()
			if err != nil {
				log.Print(err)
				continue
			}
		case "n":
			// List all bookmarks for the current song, prefixed with a number.
			for k, bm := range bms {
				fmt.Printf("%d\t%s-%s\n", k+1, bm.start, bm.end)
			}
		case "p":
			// List all bookmarks for the current song.
			for _, bm := range bms {
				fmt.Printf("%s-%s\n", bm.start, bm.end)
			}
		case "":
		default:
			fmt.Println("Unknown command")
		}
	}
}
