package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matm/bmp/pkg/config"
	"github.com/matm/bmp/pkg/mpd"
	"github.com/matm/bmp/pkg/types"
	"github.com/rotisserie/eris"
)

type shellCommand struct {
	name string
	key  string
	re   string // Regexp.
	help string
}

const exitMessage = "Bye!"
const statusBarLength = 30

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

var mu sync.Mutex

func writeBookmarks(w io.Writer, bs types.BookmarkSet) int {
	var b strings.Builder
	for song, bms := range bs {
		fmt.Fprintf(&b, "song: %s\n", song)
		for _, bm := range bms {
			fmt.Fprintf(&b, "%s-%s\n", bm.Start, bm.End)
		}
	}
	fmt.Fprintf(w, b.String())
	return b.Len()
}

var shellCmds = []shellCommand{
	{"quit", "q", `^q$`, "Exit the program"},
	{"forceQuit", "Q", `^Q$`, "Force exit the program, even with unsaved changes"},
	{"songInfo", "i", `^i$`, "Show current song information"},
	{"forward", "f", `^f$`, "Forward seek +10s in current song"},
	{"backward", "b", `^b$`, "Backward seek -10s in current song"},
	{"bookmarkStart", "[", `^\[$`, "Bookmark start: mark the beginning of the time frame"},
	{"bookmarkEnd", "]", `^\]$`, "Bookmark end: mark the end of the time frame. The time interval is added to the list of bookmarks for the current song"},
	{"deleteBookmark", "d", `^d\d*$`, "Delete bookmark entry at position pos"},
	{"deleteAllBookmarks", "D", `^D$`, "Delete all bookmark entries for current song"},
	{"change", "c", `^c(\d{1,2}) (\d{2}:\d{2})-(\d{2}:\d{2})$`, "Change bookmark entry at position pos and set new start and end time boundaries"},
	{"listBookmarks", "p", `^,?p$`, "List of current bookmarked locations in the current song"},
	{"listNumberedBookmarks", "n", `^,?n$`, "Numbered list of current bookmarked locations in the current song"},
	{"save", "w", `^w ?(.*)$`, "List bookmarks on standard output. Writes to file if argument provided"},
	{"run", "r", `^r$`, "Start the autoplay of the best parts"},
	{"stop", "s", `^s$`, "Stop the autoplay of the best parts"},
	{"toggle", "t", `^t$`, "Toggle play/pause of current song"},
	//
	{"empty", "", `^$`, ""},
	{"help", "h", `^h$`, "Show some help"},
}

func loadCommands() map[string]*regexp.Regexp {
	cmds := make(map[string]*regexp.Regexp)
	for _, cmd := range shellCmds {
		cmds[cmd.name] = regexp.MustCompile(cmd.re)
	}
	return cmds
}

func main() {
	var fname, mpdHost string
	var mpdPort int
	var showVersion bool
	flag.StringVar(&fname, "f", "", "bookmarks list file to load")
	flag.StringVar(&mpdHost, "host", os.Getenv("MPD_HOST"), "MPD host address")
	flag.IntVar(&mpdPort, "port", 6600, "MPD host TCP port")
	flag.BoolVar(&showVersion, "v", false, "show program version")
	flag.Parse()

	if showVersion {
		fmt.Printf("Version:      %s\n", config.Version)
		fmt.Printf("Git revision: %s\n", config.GitRev)
		fmt.Printf("Git branch:   %s\n", config.GitBranch)
		fmt.Printf("Go version:   %s\n", runtime.Version())
		fmt.Printf("Built:        %s\n", config.BuildDate)
		fmt.Printf("OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	if mpdHost == "" {
		fmt.Println("Missing MPD address. Please provide either $MPD_HOST or use the -host flag")
		os.Exit(2)
	}
	mp := mpd.NewClient(mpdHost, mpdPort)
	defer mp.Close()

	// Exit early if MPD doesn't reply.
	err := mp.Ping()
	if err != nil {
		fmt.Println("MPD error: connection refused")
		os.Exit(1)
	}

	quit := false
	// Keep track of bookmarks per song. The key is the song's filename.
	bms := make(types.BookmarkSet)
	// Bracket open, i.e [ for marking the beginning of a range.
	bOpen := false

	if fname != "" {
		var err error
		cf, err := os.Open(fname)
		if err != nil {
			logError(err)
			os.Exit(1)
		}
		bms, err = config.ParseBookmarkFile(cf)
		if err != nil {
			logError(eris.Wrap(err, "parsing"))
			cf.Close()
			os.Exit(1)
		}
		// Do not defer call since we're entering an infinite loop below.
		cf.Close()
		// Since a bookmark file is provided, let's load the playlist and play it
		// in auto mode.
		// Build and submit a playlist to MPD.
		ids := make([]int64, 0)
		for song := range bms {
			id, err := mp.AddToQueue(song)
			if err != nil {
				logError(eris.Wrap(err, "run cmd"))
			}
			ids = append(ids, id)
		}
		// Play first added song.
		err = mp.PlaySongID(ids[0])
		if err != nil {
			logError(err)
		}
		autoplay = true
	}

	// Start the scheduler.
	go schedule(mp, &bms)

	// Set of commands.
	cmds := loadCommands()

	// Checked before exiting.
	bufferModified := false

	p := newPrompt()

	for !quit {
		line := p.Input()
		switch {
		case cmds["help"].MatchString(line):
			for _, cmd := range shellCmds {
				if cmd.key == "" || cmd.key == "h" {
					continue
				}
				fmt.Printf("%-3s\t%s\n", cmd.key, cmd.help)
			}
		case cmds["forceQuit"].MatchString(line):
			quit = true
			fmt.Println(exitMessage)
		case cmds["quit"].MatchString(line):
			if bufferModified && len(bms) > 0 {
				fmt.Println("Warning: bookmarks list modified")
				break
			}
			quit = true
			fmt.Println(exitMessage)
		case cmds["bookmarkStart"].MatchString(line):
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
				bms[s.File] = make([]types.Bookmark, 0)
			}
			bms[s.File] = append(bms[s.File], types.Bookmark{Start: start})
			mu.Unlock()
			fmt.Println(start)
		case cmds["bookmarkEnd"].MatchString(line):
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
			bm.End = end
			mu.Unlock()
			fmt.Printf("%s-%s\n", bm.Start, bm.End)
			// Mark buffer as modified.
			bufferModified = true
		case cmds["songInfo"].MatchString(line):
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

			// Display a status bar.
			bar := makeStatusBar(statusBarLength, st.Elapsed, st.Duration, bms[s.File])
			fmt.Println(bar)
		case cmds["forward"].MatchString(line):
			// Forward seek +10s.
			err := mp.SeekOffset(10)
			if err != nil {
				log.Print(err)
				continue
			}
		case cmds["backward"].MatchString(line):
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
		case cmds["toggle"].MatchString(line):
			// Toogle play/pause.
			err := mp.Toggle()
			if err != nil {
				log.Print(err)
				continue
			}
		case cmds["listNumberedBookmarks"].MatchString(line):
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
				fmt.Printf("%d\t%s-%s\n", k+1, bm.Start, bm.End)
			}
			mu.Unlock()
		case cmds["listBookmarks"].MatchString(line):
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
				fmt.Printf("%s-%s\n", bm.Start, bm.End)
			}
			mu.Unlock()
		case cmds["save"].MatchString(line):
			// Write bookmarks buffer to stdout if no filename given.
			if len(bms) == 0 {
				fmt.Println("no bookmarks")
				break
			}
			ms := cmds["save"].FindStringSubmatch(line)
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
		case cmds["deleteBookmark"].MatchString(line):
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
				fmt.Println("missing line number (use 'n' command)")
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
		case cmds["deleteAllBookmarks"].MatchString(line):
			// Delete all bookmark entries for current song.
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
			delete(bms, s.File)
			mu.Unlock()
			// Mark buffer as modified.
			bufferModified = true
		case cmds["change"].MatchString(line):
			cs := cmds["change"].FindStringSubmatch(line)
			// Change a time range (whole line).
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
			p := cs[1]
			if p == "" {
				fmt.Println("missing line number (use 'n' command)")
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
			mu.Unlock()
			// Check start and dates are real times and end is after start.
			start, end := cs[2], cs[3]
			st, err := humanToSeconds(start)
			if err != nil {
				fmt.Printf("wrong start time format %q\n", start)
				continue
			}
			ed, err := humanToSeconds(end)
			if err != nil {
				fmt.Printf("wrong end time format %q\n", end)
				continue
			}
			if ed < st {
				fmt.Println("end time must be after start")
				continue
			}
			if st > int(s.Duration) {
				fmt.Println("start can't be greater than the song's length")
				continue
			}
			// Save new value.
			mu.Lock()
			bms[s.File][idx] = types.Bookmark{Start: start, End: end}
			mu.Unlock()
		case cmds["run"].MatchString(line):
			autoplay = true
		case cmds["stop"].MatchString(line):
			autoplay = false
		case cmds["empty"].MatchString(line):
		default:
			fmt.Println("Unknown command")
		}
	}
}
