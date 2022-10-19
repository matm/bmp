package mpd

import "testing"

func Test_tcpDialer_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"name", "MPD dialer"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &tcpDialer{}
			if got := tr.Name(); got != tt.want {
				t.Errorf("tcpDialer.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
