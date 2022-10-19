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
		name    string
		init    func() io.Reader
		wantErr bool
		after   func(err error, bs types.BookmarkSet)
	}{
		{"nil reader", func() io.Reader { return nil }, true, nil},
		{"one song but no time ranges", func() io.Reader {
			c := `
song: some/path/intro.mp3
			`
			return strings.NewReader(c)
		}, true, func(err error, bs types.BookmarkSet) {
			assert.NotNil(err)
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
