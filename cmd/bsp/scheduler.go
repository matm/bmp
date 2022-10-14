package main

import (
	"fmt"
	"time"

	"github.com/matm/bsp/pkg/mpd"
	"github.com/matm/bsp/pkg/types"
)

var autoplay = false

const schedSleep = 500 * time.Millisecond

func schedule(mp *mpd.Client, bms *types.BookmarkSet) {
	for {
		time.Sleep(schedSleep)
		if !autoplay {
			continue
		}
		// Get current song.
		// Current song info.
		s, err := mp.CurrentSong()
		if err != nil {
			// FIXME: Log error.
			time.Sleep(time.Second)
			continue
		}
		st, err := mp.Status()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		mu.Lock()
		inRange := false
		if bookmarks, ok := (*bms)[s.File]; ok {
			// This song has bookmarks.
			for _, bk := range bookmarks {
				start, err := humanToSeconds(bk.Start)
				if err != nil {
					fmt.Printf("error parsing %q", bk.Start)
					continue
				}
				if bk.End == "" {
					// A bookmark range is being defined.
					break
				}
				end, err := humanToSeconds(bk.End)
				if err != nil {
					fmt.Printf("error parsing %q", bk.End)
					continue
				}
				if int(st.Elapsed) < start && !inRange {
					err = mp.SeekTo(start)
					if err != nil {
						fmt.Printf("could not seek to %s", bk.Start)
						continue
					}
					inRange = true
				}
				if int(st.Elapsed) > end {
					inRange = false
				}
			}
		}
		mu.Unlock()
	}
}
