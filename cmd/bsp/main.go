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

func main() {
	mp := mpd.NewClient()
	defer mp.Close()
	r := bufio.NewReader(os.Stdin)

	quit := false
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
		case "m":
			fmt.Println("MARK")
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
			// Toogle.
			err := mp.Toggle()
			if err != nil {
				log.Print(err)
				continue
			}
		case "":
		default:
			fmt.Println("Unknown command")
		}
	}
}
