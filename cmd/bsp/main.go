package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

// https://mpd.readthedocs.io/en/latest/protocol.html

type mpd struct {
	conn *net.TCPConn
}
type response map[string]string

const (
	// ReplyOK is an OK reply from mpd. The command went fine.
	ReplyOK = "OK"
	// ReplyACK is mpd's way letting know there is an error.
	ReplyACK = "ACK"
)

type song struct {
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

type status struct {
	// Duration of the current song in seconds.
	Duration float64
	// Elapsed is the total time elapsed within the current song in seconds, with higher resolution.
	Elapsed float64
	SongID  string
	State   string
	Volume  int64
}

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

func (d *mpd) exec(cmd string) (response, error) {
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

func (d *mpd) CurrentSong() (*song, error) {
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
	s := &song{
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

func (d *mpd) Status() (*status, error) {
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
	s := &status{
		Duration: dur,
		Elapsed:  ela,
		SongID:   res["songid"],
		State:    res["state"],
		Volume:   vol,
	}
	return s, err
}

func (d *mpd) Stats() error {
	res, err := d.exec("stats")
	fmt.Printf("%v\n", res)
	return err
}

func (d *mpd) SeekCur() error {
	res, err := d.exec("seekcur +10")
	fmt.Printf("%v\n", res)
	return err
}

func (d *mpd) Stop() error {
	_, err := d.exec("stop")
	return err
}

func newMPD() *mpd {
	conn, err := dial()
	if err != nil {
		log.Fatal(err)
	}
	return &mpd{conn}
}

func main() {
	mp := newMPD()

	r := bufio.NewReader(os.Stdin)
	quit := false
	for !quit {
		fmt.Printf("> ")
		ch, _, err := r.ReadLine()
		if err != nil {
			log.Println(err)
		}
		switch string(ch) {
		case "q", "quit", "exit":
			quit = true
			fmt.Println("Bye!")
		case "m":
			fmt.Println("MARK")
		case "i":
			// Current song info.
			s, err := mp.CurrentSong()
			if err != nil {
				log.Print(err)
			}
			st, err := mp.Status()
			if err != nil {
				log.Print(err)
			}
			sdur, err := time.ParseDuration(fmt.Sprintf("%fs", st.Elapsed))
			if err != nil {
				log.Print(err)
			}
			fmt.Printf("[%s] %s: %s\n", st.State, s.Artist, s.Title)
			fmt.Printf("%02.2f:%02.2f\n", sdur.Minutes(), sdur.Seconds())
		case "":
		default:
			fmt.Println("Unknown command")
		}
	}
}
