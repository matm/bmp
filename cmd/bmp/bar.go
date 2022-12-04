package main

import (
	"log"
	"math"

	"github.com/matm/bmp/pkg/types"
)

func makeStatusBar(width int, elapsed, duration float64, marks []types.Bookmark) string {
	if width == 0 {
		return ""
	}
	b := make([]rune, width)
	for k := 0; k < width; k++ {
		b[k] = '-'
	}
	pos := math.Floor(elapsed * float64(width) / duration)
	for k := 0; k < int(pos); k++ {
		b[k] = '='
	}
	if pos > 1 {
		pos--
	}
	b[int(pos)] = '>'
	mu.Lock()
	for _, bm := range marks {
		start, err := humanToSeconds(bm.Start)
		if err != nil {
			mu.Unlock()
			log.Print(err)
			continue
		}
		pos := int(math.Floor(float64(start) * float64(width) / duration))
		b[pos] = '*'
	}
	mu.Unlock()
	return string(b)
}
