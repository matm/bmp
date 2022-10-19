package config

import (
	"io"
	"strings"
	"testing"

	"github.com/matm/bmp/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestParseBookmarkFile(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name  string
		init  func() io.Reader
		after func(err error, bs types.BookmarkSet)
	}{
		{"nil reader", func() io.Reader { return nil },
			func(err error, bs types.BookmarkSet) {
				assert.Error(err)
			}},
		{"one song but no time ranges", func() io.Reader {
			c := `
song: some/path/intro.mp3
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.ErrorIs(err, ErrMissingRanges)
			assert.Empty(bs)
			assert.Equal("some/path/intro.mp3: missing ranges for song, or bad time format", err.Error())
		}},
		{"2 time ranges but no song", func() io.Reader {
			c := `
01:00-01:30
02:00-02:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.ErrorIs(err, ErrOrphanRange)
			assert.Empty(bs)
			assert.Equal("[{01:00 01:30} {02:00 02:30}]: orphan ranges, missing song", err.Error())
		}},
		{"1 time range before the song", func() io.Reader {
			c := `
01:00-01:30
song: some/path/intro.mp3
02:00-02:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.ErrorIs(err, ErrOrphanRange)
			assert.Empty(bs)
		}},
		{"bad time format", func() io.Reader {
			c := `
song: some/path/intro.mp3
2:00-02:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.ErrorIs(err, ErrMissingRanges)
			assert.Empty(bs)
		}},
		{"corrupt time range", func() io.Reader {
			c := `
song: some/path/intro.mp3
2:0002:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.ErrorIs(err, ErrMissingRanges)
			assert.Empty(bs)
		}},
		{"one song, 2 ranges, 1 comment, 1 newline", func() io.Reader {
			c := `
song: some/path/intro.mp3
01:00-01:30

# Best part IMO.
02:00-02:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.NoError(err)
			assert.Len(bs, 1)
			title := "some/path/intro.mp3"
			if assert.Contains(bs, title) {
				if assert.Len(bs[title], 2) {
					assert.Equal(bs[title][0], types.Bookmark{Start: "01:00", End: "01:30"})
				}
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBookmarkFile(tt.init())
			if tt.after != nil {
				tt.after(err, got)
			}
		})
	}
}
