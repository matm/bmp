package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matm/bsp/pkg/mpd"
	"github.com/matm/bsp/pkg/types"
	"github.com/rotisserie/eris"
)

func logError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}

func secondsToHuman(secs int) string {
	m := secs / 60
	s := secs % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func humanToSeconds(t string) (int, error) {
	// FIXME: does not support hours-long songs.
	p, err := time.Parse("04:05", t)
	if err != nil {
		return -1, eris.Wrap(err, "human2seconds")
	}
	return p.Hour()*3600 + p.Minute()*60 + p.Second(), nil
}

type bookmark struct {
	// Both start and end have MM:SS formatting.
	start, end string
}

type bookmarkSet map[string][]bookmark

var mu sync.Mutex

func writeBookmarks(w io.Writer, bs bookmarkSet) int {
	var b strings.Builder
	for song, bms := range bs {
		fmt.Fprintf(&b, "song: %s\n", song)
		for _, bm := range bms {
			fmt.Fprintf(&b, "%s-%s\n", bm.start, bm.end)
		}
	}
	fmt.Fprintf(w, b.String())
	return b.Len()
}

func schedule(mp *mpd.Client, bms *bookmarkSet) {
	for {
		// Get current song.
		// Current song info.
		s, err := mp.CurrentSong()
		if err != nil {
			// FIXME: Log error.
			time.Sleep(time.Second)
			continue
		}
		/*
			st, err := mp.Status()
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
		*/
		mu.Lock()
		if bookmarks, ok := (*bms)[s.File]; ok {
			for _, bk := range bookmarks {
				to, err := humanToSeconds(bk.start)
				if err != nil {
					fmt.Printf("error parsing %q", bk.start)
					continue
				}
				err = mp.SeekTo(to)
				if err != nil {
					fmt.Printf("could not seek to %s", bk.start)
					continue
				}
			}
		}
		mu.Unlock()
		time.Sleep(time.Second)
	}
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
	cmdSave := regexp.MustCompile(`^w ?(.*)$`)
	cmdDeleteBookmark := regexp.MustCompile(`^d\d*$`)
	cmdEmpty := regexp.MustCompile("^$")

	// Run the scheduler.
	// BUGGY FOR NOW
	//go schedule(mp, &bms)

	// Checked before exiting.
	bufferModified := false

	for !quit {
		fmt.Printf("> ")
		ch, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
		}
		line := string(ch)
		switch {
		case cmdQuit.MatchString(line):
			if bufferModified {
				fmt.Println("Warning: bookmarks list modified")
				break
			}
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
			mu.Lock()
			if _, ok := bms[s.File]; !ok {
				bms[s.File] = make([]bookmark, 0)
			}
			bms[s.File] = append(bms[s.File], bookmark{start: start})
			mu.Unlock()
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
			mu.Lock()
			bm := &bms[s.File][len(bms[s.File])-1]
			bm.end = end
			mu.Unlock()
			fmt.Printf("%s-%s\n", bm.start, bm.end)
			// Mark buffer as modified.
			bufferModified = true
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
			mu.Lock()
			if _, ok := bms[s.File]; !ok {
				mu.Unlock()
				continue
			}
			for k, bm := range bms[s.File] {
				fmt.Printf("%d\t%s-%s\n", k+1, bm.start, bm.end)
			}
			mu.Unlock()
		case cmdListBookmarks.MatchString(line):
			// List all bookmarks for the current song.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			mu.Lock()
			if _, ok := bms[s.File]; !ok {
				mu.Unlock()
				continue
			}
			for _, bm := range bms[s.File] {
				fmt.Printf("%s-%s\n", bm.start, bm.end)
			}
			mu.Unlock()
		case cmdSave.MatchString(line):
			// Write bookmarks buffer to stdout if no filename given.
			ms := cmdSave.FindStringSubmatch(line)
			filename := ms[len(ms)-1]
			if filename == "" {
				writeBookmarks(os.Stdout, bms)
				break
			}
			persist := func() error {
				f, err := os.Create(filename)
				if err != nil {
					return eris.Wrap(err, "save bookmark file")
				}
				defer f.Close()
				fmt.Println(writeBookmarks(f, bms))
				bufferModified = false
				return nil
			}
			if err := persist(); err != nil {
				logError(err)
				break
			}
			break
		case cmdDeleteBookmark.MatchString(line):
			// Delete a bookmark entry for current song.
			// Bookmark ID to delete starts at 1.
			s, err := mp.CurrentSong()
			if err != nil {
				if err != types.ErrNoSong {
					log.Print(err)
				}
				continue
			}
			mu.Lock()
			if _, ok := bms[s.File]; !ok {
				fmt.Println("no bookmark for this song")
				mu.Unlock()
				continue
			}
			mu.Unlock()
			p := line[1:]
			if p == "" {
				fmt.Println("missing bookmark entry number (use 'n' command)")
				continue
			}
			idx, err := strconv.ParseInt(p, 10, 64)
			if err != nil {
				log.Print(err)
				continue
			}
			idx--
			mu.Lock()
			if int(idx) > len(bms[s.File])-1 || idx < 0 {
				fmt.Printf("out of range\n")
				mu.Unlock()
				continue
			}
			bms[s.File] = append(bms[s.File][:int(idx)], bms[s.File][int(idx)+1:]...)
			mu.Unlock()
			// Mark buffer as modified.
			bufferModified = true
		case cmdEmpty.MatchString(line):
		default:
			fmt.Println("Unknown command")
		}
	}
}
