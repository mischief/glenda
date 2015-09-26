// +build ignore

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/kballard/goirc/irc"
)

var shorten = false

// maps a feed to a set of channels to send feed results to
type feedspitter struct {
	conn     irc.SafeConn
	feed     *rss.Feed
	name     string
	color    string
	freq     time.Duration
	seen     map[string]bool
	channels []string
	stop     chan bool
}

// newfeedsplitter constructs a new feedsplitter
func newfeedsplitter(conn irc.SafeConn, url, name, color string, freq time.Duration, channels []string) (*feedspitter, error) {
	f, err := rss.Fetch(url)

	if err != nil {
		return nil, err
	}

	fs := &feedspitter{
		conn:     conn,
		feed:     f,
		name:     name,
		color:    color,
		freq:     freq,
		seen:     make(map[string]bool),
		channels: channels,
		stop:     make(chan bool),
	}

	// initial sweep of seen items
	for _, i := range fs.feed.Items {
		fs.seen[i.ID] = true
	}

	return fs, nil
}

// update feeds, send new items to irc channels
func (f *feedspitter) update() {
	for {
		select {
		case <-f.stop:
			return
		default:
			if err := f.feed.Update(); err != nil {
				log.Printf("aborting %s: %s", f.feed.UpdateURL, err)
			}

			//log.Printf("%s unread: %d update: %s", f.feed.UpdateURL, f.feed.Unread, f.feed.Refresh)

			if f.feed.Unread > 0 {

				for _, i := range f.feed.Items {
					if _, ok := f.seen[i.ID]; !ok {
						var url string
						/* check for url shortener */
						if m := GetModule("urlshortener"); m != nil && shorten == true {
							sh := m.(*UrlShortenerMod)
							u, err := sh.shorten(i.Link)
							if err != nil {
								url = err.Error()
							} else {
								url = u
							}
						} else {
							url = i.Link
						}
						for _, ch := range f.channels {
							var name string
							if f.name != "" {
								name = fmt.Sprintf("%s: ", f.name)
							}
							f.conn.Privmsg(ch, Colored(fmt.Sprintf("%s%s: %s", name, i.Title, url), f.color))
						}
						f.seen[i.ID] = true
					}
				}

				f.feed.Unread = 0
				if f.freq >= 0 {
					f.feed.Refresh = time.Now().Add(f.freq)
				}

			}

			time.Sleep(90 * time.Second)

		}
	}
}

type FeedReaderMod struct {
	feeds map[string]*feedspitter
}

func (f *FeedReaderMod) Init(b *Bot, conn irc.SafeConn) (err error) {
	f.feeds = make(map[string]*feedspitter)

	conf := b.Config.Search("mod", "feedreader")

	shortens := conf.Search("shorten")
	if shortens == "true" || shortens == "1" {
		shorten = true
	}

	for _, rec := range conf {
		for _, line := range rec {
			var url string
			var channels []string
			var freqs string
			var name string
			var color string

			for _, tup := range line {
				if tup.Attr == "feed" {
					url = tup.Val
				}
				if tup.Attr == "channels" {
					channels = strings.Fields(tup.Val)
				}
				if tup.Attr == "freq" {
					freqs = tup.Val
				}
				if tup.Attr == "color" {
					color = tup.Val
				}
				if tup.Attr == "name" {
					name = tup.Val
				}
			}

			if url != "" && channels != nil && len(channels) > 0 {
				// allow frequency to be unset; this means use the rss reader default
				var freq time.Duration
				if freqs != "" {
					pfreq, err := time.ParseDuration(freqs)
					if err != nil {
						log.Printf("%s skipped: %s", url, err)
						continue
					}
					freq = pfreq
				}

				if color == "" {
					color = "white"
				}

				fs, err := newfeedsplitter(conn, url, name, color, freq, channels)
				if err != nil {
					log.Printf("%s skipped: %s", url, err)
					continue
				}

				f.feeds[url] = fs
				log.Printf("added feed %s", url)
			}
		}
	}

	for _, fs := range f.feeds {
		go fs.update()
	}

	log.Printf("feedreader module initialized")

	return nil
}
