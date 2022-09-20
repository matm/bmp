package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/matm/bsp/pkg/mpd"
	"github.com/matm/bsp/pkg/types"
)

func secondsToHuman(secs int) string {
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

type bookmark struct {
	// Both start and end have MM:SS formatting.
	start, end string
}

type bookmarkSet map[string][]bookmark

func save(w io.Writer, bs bookmarkSet) error {
	for song, bms := range bs {
		fmt.Fprintf(w, "song: %s\n", song)
		for _, bm := range bms {
			fmt.Fprintf(w, "%s-%s\n", bm.start, bm.end)
		}
	}
	return nil
}

func main() {
	mp := mpd.NewClient()
	defer mp.Close()
	r := bufio.NewReader(os.Stdin)

	quit := false
	// Keep track of bookmarks per song. The key is the song's filename.
	bms := make(bookmarkSet)
	bOpen := false

	cmdQuit := regexp.MustCompile(`^q$`)
	cmdBookmarkStart := regexp.MustCompile(`^\[$`)
	cmdBookmarkEnd := regexp.MustCompile(`^\]$`)
	cmdSongInfo := regexp.MustCompile(`^i$`)
	cmdForward := regexp.MustCompile(`^f$`)
	cmdBackward := regexp.MustCompile(`^b$`)
	cmdToggle := regexp.MustCompile(`^t$`)
	cmdListBookmarks := regexp.MustCompile(`^p$`)
	cmdListNumberedBookmarks := regexp.MustCompile(`^n$`)
	cmdSave := regexp.MustCompile(`^w$`)
	cmdDeleteBookmark := regexp.MustCompile(`^d\d*$`)
	cmdEmpty := regexp.MustCompile("^$")

	for !quit {
		fmt.Printf("> ")
		ch, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
		}
		line := string(ch)
		switch {
		case cmdQuit.MatchString(line):
			quit = true
			fmt.Println("Bye!")
		case cmdBookmarkStart.MatchString(line):
			// Bookmark start.
			st, err := mp.Status()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			if st.State != "play" {
				fmt.Println("Please starting playing a song first")
				continue
			}
			if bOpen {
				fmt.Println("Missing closing bookmark, please use ']' first")
				continue
			}
			// Current song info.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			bOpen = true
			start := secondsToHuman(int(st.Elapsed))
			if _, ok := bms[s.File]; !ok {
				bms[s.File] = make([]bookmark, 0)
			}
			bms[s.File] = append(bms[s.File], bookmark{start: start})
			fmt.Println(start)
		case cmdBookmarkEnd.MatchString(line):
			// Bookmark end.
			st, err := mp.Status()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			if st.State != "play" {
				fmt.Println("Please starting playing a song first")
				continue
			}
			if !bOpen {
				fmt.Println("Missing opening bookmark, please use '[' first")
				continue
			}
			// Current song info.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			bOpen = false
			end := secondsToHuman(int(st.Elapsed))
			bm := &bms[s.File][len(bms[s.File])-1]
			bm.end = end
			fmt.Printf("%s-%s\n", bm.start, bm.end)
		case cmdSongInfo.MatchString(line):
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
		case cmdForward.MatchString(line):
			// Forward seek +10s.
			err := mp.SeekOffset(10)
			if err != nil {
				log.Print(err)
				continue
			}
		case cmdBackward.MatchString(line):
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
		case cmdToggle.MatchString(line):
			// Toogle play/pause.
			err := mp.Toggle()
			if err != nil {
				log.Print(err)
				continue
			}
		case cmdListNumberedBookmarks.MatchString(line):
			// List all bookmarks for the current song, prefixed with a number.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			if _, ok := bms[s.File]; !ok {
				continue
			}
			for k, bm := range bms[s.File] {
				fmt.Printf("%d\t%s-%s\n", k+1, bm.start, bm.end)
			}
		case cmdListBookmarks.MatchString(line):
			// List all bookmarks for the current song.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			if _, ok := bms[s.File]; !ok {
				continue
			}
			for _, bm := range bms[s.File] {
				fmt.Printf("%s-%s\n", bm.start, bm.end)
			}
		case cmdSave.MatchString(line):
			// Persist to disk.
			err := save(os.Stdout, bms)
			if err != nil {
				log.Print(err)
			}
		case cmdDeleteBookmark.MatchString(line):
			// Delete a bookmark entry for current song.
			// Bookmark ID to delete starts at 1.
			if len(bms) == 0 {
				continue
			}
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			p := line[1:]
			if p == "" {
				fmt.Println("missing bookmark entry number")
				continue
			}
			idx, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				log.Print(err)
				continue
			}
			idx--
			if int(idx) > len(bms[s.File])-1 || idx <= 0 {
				fmt.Printf("out of range\n")
				continue
			}
			bms[s.File] = append(bms[s.File][:int(idx)], bms[s.File][int(idx)+1:]...)
		case cmdEmpty.MatchString(line):
		default:
			fmt.Println("Unknown command")
		}
	}
}
