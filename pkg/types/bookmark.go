package types

// Bookmark is a time range (point of interest), with a start and end time.
// Both start and end have MM:SS formatting.
type Bookmark struct {
	Start, End string
}

// BookmarkSet is a map of song name as a key and an associated list of bookmarks.
type BookmarkSet map[string][]Bookmark
