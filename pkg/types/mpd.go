package types

import "errors"

// ErrNoSong returned when no current song is active.
var ErrNoSong = errors.New("no current song")

// Song info from MPD.
type Song struct {
	ID           string
	Album        string
	Artist       string
	Date         string
	Duration     float64
	File         string
	Genre        string
	LastModified string
	Pos          int64
	Time         int64
	Title        string
	Track        string
}

// Status of current song.
type Status struct {
	// Duration of the current song in seconds.
	Duration float64
	// Elapsed is the total time elapsed within the current song in seconds, with higher resolution.
	Elapsed float64
	SongID  string
	State   string
	Volume  int64
}
