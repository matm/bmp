package main

import (
	"fmt"
	"time"

	"github.com/matm/bmp/pkg/mpd"
	"github.com/matm/bmp/pkg/types"
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
		bookmarks, ok := (*bms)[s.File]
		mu.Unlock()
		if !ok {
			continue
		}
		// This song has bookmarks.
		for j, bk := range bookmarks {
			start, err := humanToSeconds(bk.Start)
			if err != nil {
				fmt.Printf("error parsing %q", bk.Start)
				continue
			}
			if bk.End == "" {
				// A bookmark range is being defined.
				break
			}
			if int(st.Elapsed) < start {
				canSeek := false
				if j == 0 {
					canSeek = true
				} else {
					end, err := humanToSeconds(bookmarks[j-1].End)
					if err != nil {
						fmt.Printf("error parsing %q", bk.End)
						continue
					}
					if int(st.Elapsed) > end {
						canSeek = true
					}
				}
				if canSeek {
					err = mp.SeekTo(start)
					if err != nil {
						fmt.Printf("could not seek to %s", bk.Start)
						continue
					}
				}
			}
		}
	}
}
