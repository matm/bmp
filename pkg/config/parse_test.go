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
			assert.Error(err)
			assert.Empty(bs)
		}},
		{"2 time ranges but no song", func() io.Reader {
			c := `
01:00-01:30
02:00-02:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.Error(err)
			assert.Empty(bs)
		}},
		{"1 time range before the song", func() io.Reader {
			c := `
01:00-01:30
song: some/path/intro.mp3
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.Error(err)
			assert.Empty(bs)
		}},
		{"bad time format", func() io.Reader {
			c := `
song: some/path/intro.mp3
2:00-02:30
			`
			return strings.NewReader(c)
		}, func(err error, bs types.BookmarkSet) {
			assert.Error(err)
			assert.Empty(bs)
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
