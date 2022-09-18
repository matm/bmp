package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

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
	fmt.Fprintf(d.conn, "%s\n", cmd)
	sc := bufio.NewScanner(d.conn)

	resp := make(response)

	for sc.Scan() {
		line := sc.Text()
		if line == ReplyOK {
			break
		}
		if strings.HasPrefix(line, ReplyACK) {
			// Some error.
			return nil, eris.New(line)
		}
		sp := strings.Split(line, ": ")
		if len(sp) == 2 {
			// This is a key: value response line.
			resp[sp[0]] = sp[1]
		}
	}
	err := sc.Err()
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
			err := mp.CurrentSong()
			if err != nil {
				log.Print(err)
			}
			err = mp.Status()
			if err != nil {
				log.Print(err)
			}
		case "":
		default:
			fmt.Println("Unknown command")
		}
	}
}
