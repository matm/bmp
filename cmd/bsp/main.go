package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/rotisserie/eris"
)

// https://mpd.readthedocs.io/en/latest/protocol.html

type mpd struct{}
type response map[string]string

func (d *mpd) dial() (net.Conn, error) {
	return net.DialTimeout("tcp", "musik:6600", 5*time.Second)
}

func (d *mpd) exec(cmd string) (response, error) {
	conn, err := d.dial()
	if err != nil {
		return nil, eris.Wrap(err, "dial")
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\n", cmd)
	sc := bufio.NewScanner(conn)

	resp := make(response)

	for sc.Scan() {
		line := sc.Text()
		if line == "OK" {
			break
		}
		if strings.HasPrefix(line, "ACK") {
			// Some error.
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

func (d *mpd) CurrentSong() error {
	res, err := d.exec("currentsong")
	fmt.Printf("%v\n", res)
	return err
}

func (d *mpd) Status() error {
	res, err := d.exec("status")
	fmt.Printf("%v\n", res)
	return err
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
	return &mpd{}
}

func main() {
	// TODO: read $MPD_HOST
	s := newMPD()
	if err := s.Status(); err != nil {
		log.Fatal(err)
	}
}
