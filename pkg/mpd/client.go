package mpd

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/matm/bsp/pkg/types"
	"github.com/rotisserie/eris"
)

// Client to MPD.
// Doc at https://mpd.readthedocs.io/en/latest/protocol.html.
type Client struct {
	conn net.Conn
	dial dialer
}

type response map[string]string

type commander interface {
	Exec(cmd string) (response, error)
}

const (
	// ReplyOK is an OK reply from mpd. The command went fine.
	ReplyOK = "OK"
	// ReplyACK is mpd's way letting know there is an error.
	ReplyACK = "ACK"
)

func (d *Client) exec(cmd string) (response, error) {
	if d.conn == nil {
		conn, err := d.dial.Dial()
		if err != nil {
			return nil, eris.Wrap(err, "dial")
		}
		d.conn = conn
	}
	_, err := d.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		// Reconnect in case of broken pipe error.
		if errors.Is(err, syscall.EPIPE) {
			d.conn.Close()
			conn, err := d.dial.Dial()
			if err != nil {
				return nil, eris.Wrap(err, "(re)dial")
			}
			d.conn = conn
			_, err = d.conn.Write([]byte(cmd + "\n"))
			if err != nil {
				return nil, eris.Wrap(err, "exec")
			}
		} else {
			return nil, eris.Wrap(err, "exec")
		}
	}

	sc := bufio.NewScanner(d.conn)
	resp := make(response)

	for sc.Scan() {
		line := sc.Text()
		if line == ReplyOK {
			break
		}
		if strings.HasPrefix(line, ReplyACK) {
			return nil, eris.New(line)
		}
		sp := strings.Split(line, ": ")
		if len(sp) == 2 {
			// This is a key: value response line.
			resp[sp[0]] = sp[1]
		}
	}
	err = sc.Err()
	if err != nil {
		return nil, eris.Wrap(err, "scan")
	}
	return resp, nil
}

func (d *Client) CurrentSong() (*types.Song, error) {
	res, err := d.exec("currentsong")
	if err != nil {
		return nil, eris.Wrap(err, "current song")
	}
	if res["Id"] == "" {
		// Empty reply, no current song.
		return nil, types.ErrNoSong
	}
	dur, err := strconv.ParseFloat(res["duration"], 64)
	if err != nil {
		return nil, eris.Wrap(err, "current song: duration")
	}
	pos, err := strconv.ParseInt(res["Pos"], 10, 64)
	if err != nil {
		return nil, eris.Wrap(err, "current song: pos")
	}
	ti, err := strconv.ParseInt(res["Time"], 10, 64)
	if err != nil {
		return nil, eris.Wrap(err, "current song: time")
	}
	s := &types.Song{
		ID:           res["Id"],
		Album:        res["Album"],
		Artist:       res["Artist"],
		Date:         res["Date"],
		Duration:     dur,
		File:         res["file"],
		Genre:        res["Genre"],
		LastModified: res["Last-Modified"],
		Pos:          pos,
		Time:         ti,
		Title:        res["Title"],
		Track:        res["Track"],
	}
	return s, err
}

func (d *Client) Status() (*types.Status, error) {
	res, err := d.exec("status")
	if err != nil {
		return nil, eris.Wrap(err, "status")
	}
	if res["duration"] == "" {
		// Empty reply, no current song.
		return nil, types.ErrNoSong
	}
	dur, err := strconv.ParseFloat(res["duration"], 64)
	if err != nil {
		return nil, eris.Wrap(err, "status: duration")
	}
	ela, err := strconv.ParseFloat(res["elapsed"], 64)
	if err != nil {
		return nil, eris.Wrap(err, "status: elapsed")
	}
	vol, err := strconv.ParseInt(res["volume"], 10, 64)
	if err != nil {
		return nil, eris.Wrap(err, "status: volume")
	}
	s := &types.Status{
		Duration: dur,
		Elapsed:  ela,
		SongID:   res["songid"],
		State:    res["state"],
		Volume:   vol,
	}
	return s, err
}

func (d *Client) Stats() error {
	res, err := d.exec("stats")
	fmt.Printf("%v\n", res)
	return eris.Wrap(err, "stats")
}

// Toggle pauses or resumes playback. The pause state is toggled.
func (d *Client) Toggle() error {
	_, err := d.exec("pause")
	return eris.Wrap(err, "toggle")
}

func (d *Client) SeekOffset(offset int) error {
	sig := "+"
	if offset < 0 {
		sig = ""
	}
	_, err := d.exec(fmt.Sprintf("seekcur %s%d", sig, offset))
	return err
}

func (d *Client) SeekTo(seconds int) error {
	_, err := d.exec(fmt.Sprintf("seekcur %d", seconds))
	return err
}

func (d *Client) Stop() error {
	_, err := d.exec("stop")
	return err
}

func NewClient() *Client {
	return &Client{dial: defaultDialer}
	//return &Client{dial: testDialer}
}

// Close terminates the TCP connection.
func (d *Client) Close() error {
	var err error
	if d.conn != nil {
		err = d.conn.Close()
	}
	return err
}
