package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/matm/bsp/pkg/mpd"
)

func main() {
	mp := mpd.NewClient()
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
				log.Print(err)
			}
			st, err := mp.Status()
			if err != nil {
				log.Print(err)
			}
			sdur, err := time.ParseDuration(fmt.Sprintf("%fs", st.Elapsed))
			if err != nil {
				log.Print(err)
			}
			fmt.Printf("[%s] %s: %s\n", st.State, s.Artist, s.Title)
			fmt.Printf("%02.2f:%02.2f\n", sdur.Minutes(), sdur.Seconds())
		case "":
		default:
			fmt.Println("Unknown command")
		}
	}
}
