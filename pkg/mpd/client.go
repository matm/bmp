package mpd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/matm/bsp/pkg/types"
	"github.com/rotisserie/eris"
)

// Client to MPD.
// Doc at https://mpd.readthedocs.io/en/latest/protocol.html.
type Client struct {
	conn *net.TCPConn
}
type response map[string]string

const (
	// ReplyOK is an OK reply from mpd. The command went fine.
	ReplyOK = "OK"
	// ReplyACK is mpd's way letting know there is an error.
	ReplyACK = "ACK"
)

func dial() (*net.TCPConn, error) {
	ips, err := net.LookupIP(os.Getenv("MPD_HOST"))
	if err != nil {
		return nil, eris.Wrap(err, "dial")
	}
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   ips[0],
		Port: 6600,
	})
	if err != nil {
		return nil, eris.Wrap(err, "dial")
	}
	// Keep the TCP connection alive.
	conn.SetKeepAlive(true)
	return conn, nil
}

func (d *Client) exec(cmd string) (response, error) {
	_, err := d.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return nil, eris.Wrap(err, "write to connection")
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
	dur, err := strconv.ParseFloat(res["duration"], 64)
	if err != nil {
		return nil, eris.Wrap(err, "current song: duration")
	}
	ela, err := strconv.ParseFloat(res["elapsed"], 64)
	if err != nil {
		return nil, eris.Wrap(err, "current song: elapsed")
	}
	vol, err := strconv.ParseInt(res["volume"], 10, 64)
	if err != nil {
		return nil, eris.Wrap(err, "current song: volume")
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
	return err
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
	conn, err := dial()
	if err != nil {
		log.Fatal(err)
	}
	return &Client{conn}
}

// Close terminates the TCP connection.
func (d *Client) Close() error {
	return d.conn.Close()
}
