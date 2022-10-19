package config

import (
	"bufio"
	"fmt"
	"io"
	"regexp"

	"github.com/matm/bmp/pkg/types"
	"github.com/rotisserie/eris"
)

// ParseBookmarkFile reads a bookmarks file and loads all bookmark entries.
func ParseBookmarkFile(r io.Reader) (types.BookmarkSet, error) {
	if r == nil {
		return nil, eris.New("nil reader")
	}
	bms := make(types.BookmarkSet)

	// File format is
	// song: mpd_relative_path_to_song.mp3
	// time_start:time_end
	// time_start:time_end
	// # This is a comment. Will be ignored.
	// song: another_song.flac
	// time_start:time_end
	// ...
	// Example:
	// song: metal/Metallica/BlackAlbum/the_unforgiven.mp3
	// 01:02-01:03
	// 01:34-02:12

	songRE := regexp.MustCompile(`^song: *(.*)$`)
	commentRE := regexp.MustCompile(`^#`)
	timeRE := regexp.MustCompile(`^([0-9]{2}:[0-9]{2})-([0-9]{2}:[0-9]{2})`)

	sc := bufio.NewScanner(r)
	numSongs, numBookmarks := 0, 0
	var songName string
	for sc.Scan() {
		line := sc.Text()
		switch {
		case songRE.MatchString(line):
			sn := songRE.FindStringSubmatch(line)
			songName = sn[len(sn)-1]
			if _, ok := bms[songName]; !ok {
				bms[songName] = make([]types.Bookmark, 0)
			}
			numSongs++
		case timeRE.MatchString(line):
			times := timeRE.FindStringSubmatch(line)
			bk := types.Bookmark{
				Start: times[len(times)-2],
				End:   times[len(times)-1],
			}
			bms[songName] = append(bms[songName], bk)
			numBookmarks++
		case commentRE.MatchString(line):
		default:
		}
	}
	err := sc.Err()
	if err != nil {
		return nil, eris.Wrap(err, "bookmark scan")
	}
	fmt.Printf("Loaded %d songs, %d bookmarks\n", numSongs, numBookmarks)
	return bms, nil
}
