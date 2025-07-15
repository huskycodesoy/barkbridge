package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

// Event represents an SMS or call event to POST.
type Event struct {
	Type      string ` + "`json:\"type\"`" + `
	Modem     string ` + "`json:\"modem\"`" + `
	Timestamp string ` + "`json:\"timestamp\"`" + `
	From      string ` + "`json:\"from\"`" + `
	Text      string ` + "`json:\"text,omitempty\"`" + `
}

var (
	endpoint = flag.String("endpoint", "", "POST endpoint URL")
	baud     = flag.Int("baud", 115200, "Serial baud rate")
)

type Modem struct {
	Path   string
	Port   *serial.Port
	rw     *bufio.ReadWriter
	stopCh chan struct{}
}

var (
	modemsMu sync.Mutex
	modems   = map[string]*Modem{}
)

func main() {
	flag.Parse()
	if *endpoint == "" {
		log.Fatal("endpoint flag required")
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	scanModems()
	for range ticker.C {
		scanModems()
	}
}

func scanModems() {
	paths, err := filepath.Glob("/dev/ttyUSB*")
	if err != nil {
		log.Println("glob error:", err)
		return
	}

	modemsMu.Lock()
	defer modemsMu.Unlock()

	for _, p := range paths {
		if _, ok := modems[p]; !ok {
			if e := attachModem(p); e != nil {
				log.Println("attach error", p, e)
			}
		}
	}

	for p, m := range modems {
		found := false
		for _, q := range paths {
			if q == p {
				found = true
				break
			}
		}
		if !found {
			m.stop()
			delete(modems, p)
			log.Println("detached", p)
		}
	}
}

func attachModem(path string) error {
	c := &serial.Config{Name: path, Baud: *baud, ReadTimeout: 5 * time.Second}
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}

	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	m := &Modem{Path: path, Port: s, rw: rw, stopCh: make(chan struct{})}
	modems[path] = m
	go m.run()
	return nil
}

func (m *Modem) run() {
	m.sendCmd("AT+CMGF=1")
	m.sendCmd("AT+CNMI=2,1,0,0,0")
	m.sendCmd("AT+CLIP=1")

	scanner := bufio.NewScanner(m.rw.Reader)
	for {
		select {
		case <-m.stopCh:
			m.Port.Close()
			return
		default:
			if !scanner.Scan() {
				continue
			}
			m.handleLine(scanner.Text())
		}
	}
}

func (m *Modem) sendCmd(cmd string) {
	m.rw.WriteString(cmd + "\r")
	m.rw.Flush()
}

func (m *Modem) handleLine(line string) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "+CMT:") {
		parts := strings.Split(line, ",")
		from := strings.Trim(parts[1], "\"")
		if buf, err := m.readNextLine(); err == nil {
			ev := Event{Type: "sms", Modem: m.Path, From: from, Text: buf, Timestamp: time.Now().Format(time.RFC3339)}
			sendEvent(ev)
		}
	} else if strings.HasPrefix(line, "+CLIP:") {
		parts := strings.SplitN(line, ",", 2)
		num := strings.Trim(parts[0][7:], "\"")
		ev := Event{Type: "call", Modem: m.Path, From: num, Timestamp: time.Now().Format(time.RFC3339)}
		sendEvent(ev)
	}
}

func (m *Modem) readNextLine() (string, error) {
	line, err := m.rw.Reader.ReadString('\n')
	return strings.TrimSpace(line), err
}

func (m *Modem) stop() {
	close(m.stopCh)
}

func sendEvent(ev Event) {
	b, _ := json.Marshal(ev)
	resp, err := http.Post(*endpoint, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Println("post err", err)
		return
	}
	ioutil.ReadAll(resp.Body)
	resp.Body.Close()
}
