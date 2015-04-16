package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/kballard/goirc/irc"
	"github.com/robfig/cron"
	"io"
	"log"
	"os"
	"time"
)

func init() {
	RegisterModule("wtmp", func() Module {
		return &WtmpMod{}
	})
}

type WtmpMod struct {
	file string
	last time.Time
	cron *cron.Cron
	// tty -> name
	on map[string]string
}

type WtmpEntry struct {
	line, name, host string
	date             time.Time
}

func wtmp(path string) ([]WtmpEntry, error) {
	var ent []WtmpEntry

	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	br := bufio.NewReader(f)

	for {
		line := make([]byte, 8)
		name := make([]byte, 32)
		host := make([]byte, 256)
		var date int64

		if _, err = io.ReadFull(br, line); err != nil {
			break
		}
		if _, err = io.ReadFull(br, name); err != nil {
			break
		}
		if _, err = io.ReadFull(br, host); err != nil {
			break
		}
		if err = binary.Read(br, binary.LittleEndian, &date); err != nil {
			break
		}

		entry := WtmpEntry{
			line: string(bytes.TrimRight(line, "\x00")),
			name: string(bytes.TrimRight(name, "\x00")),
			host: string(bytes.TrimRight(host, "\x00")),
			date: time.Unix(date, 0),
		}

		ent = append(ent, entry)
	}

	return ent, nil
}

func (w *WtmpMod) Init(b *Bot, conn irc.SafeConn) error {
	conf := b.Config.Search("mod", "wtmp")
	w.file = conf.Search("file")
	w.last = time.Now()
	w.cron = cron.New()
	w.on = make(map[string]string)

	channel := conf.Search("channel")

	w.cron.AddFunc("@every 1m", func() {
		//log.Printf("checking wtmp %s...", w.file)

		wtmps, err := wtmp(w.file)
		if err != nil {
			log.Printf("error checking wtmp: %s", err)
			return
		}

		for _, wtr := range wtmps {
			if wtr.name != "" {
				w.on[wtr.line] = wtr.name
			}

			if w.last.Before(wtr.date) {
				log.Printf("wtmp: %q %q %q %q", wtr.line, wtr.name, wtr.host, wtr.date)

				in := "in "
				if wtr.name == "" {
					in = "out"
					wtr.name = w.on[wtr.line]
				}
				conn.Privmsg(channel, fmt.Sprintf("log%s: %s on %s", in, wtr.name, wtr.line))
			}
		}

		w.last = time.Now()
	})

	go func() {
		time.Sleep(10 * time.Second)
		w.cron.Start()
	}()

	log.Printf("wtmp module initialized with file %s", w.file)

	return nil
}

func (m *WtmpMod) Reload() error {
	return nil
}

func (m *WtmpMod) Call(args ...string) error {
	return nil
}
