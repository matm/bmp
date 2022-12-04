package main

import (
	"testing"

	"github.com/matm/bmp/pkg/types"
)

func Test_makeStatusBar(t *testing.T) {
	type args struct {
		width    int
		elapsed  float64
		duration float64
		marks    []types.Bookmark
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"no width", args{0, 15.0, 30.0, nil}, ""},
		{"beginning", args{10, 0.0, 30.0, nil}, ">---------"},
		{"half", args{10, 15.0, 30.0, nil}, "====>-----"},
		{"full", args{10, 30.0, 30.0, nil}, "=========>"},
		{"current pos before bookmark", args{10, 10.0, 30.0, []types.Bookmark{{Start: "00:15", End: "00:20"}}}, "==>--*----"},
		{"current pos after bookmark", args{10, 25.0, 30.0, []types.Bookmark{{Start: "00:15", End: "00:30"}}}, "=====*=>--"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeStatusBar(tt.args.width, tt.args.elapsed, tt.args.duration, tt.args.marks); got != tt.want {
				t.Errorf("makeStatusBar() = %v, want %v", got, tt.want)
			}
		})
	}
}
