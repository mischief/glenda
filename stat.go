package main

import (
	"database/sql"
	_"github.com/mattn/go-sqlite3"

	"github.com/kballard/goirc/irc"

	"log"

	"bufio"
	"strings"
	"unicode/utf8"
	"unicode"
	"fmt"
)

type StatMod struct {
	db *sql.DB
}

var tables = []string{
	`CREATE TABLE IF NOT EXISTS
		Stat (
			id INTEGER PRIMARY KEY autoincrement,
			ident TEXT,

			chars INTEGER,
			words INTEGER,
			sentences INTEGER,
			lines INTEGER
		)`,
}

func init () {
	RegisterModule("stat", func () Module {
		return &StatMod{}
	})
}

func (m *StatMod) Init (b *Bot, conn irc.SafeConn) (err error) {
	conf := b.Config.Search("mod", "stat")

	m.db, err = sql.Open("sqlite3", conf.Search("path"))

	if err != nil {
		log.Printf("stat module failed to open %q: %s\n", conf.Search("path"), err)
		return
	}

	for _, t := range tables {
		_, err = m.db.Exec(t)

		if err != nil {
			log.Printf("stat module failed to create table: %s\n%q\n", err, t)
			return
		}
	}

	conn.AddHandler("PRIVMSG", func (c *irc.Conn, l irc.Line) {
		args := strings.Split(l.Args[1], " ")

		if len(args) >= 1 && args[0] == ".stat" {

			var recipient string

			if l.Args[0][0] == '#'{
				recipient = l.Args[0]
			} else {
				recipient = l.Src.String()
			}

			if len(args) >= 2 {
				s, err := m.stat(args[1])

				if err != nil {
					log.Printf("stat failed for %q: %s", args[1], err)
					return
				}

				if s == nil {
					c.Privmsg(recipient, `no such user: "` + args[1] + `"`)
					return
				}
				
				c.Privmsg(recipient, s.String())
			} else {
				c.Privmsg(recipient, "missing argument")
			}
			
		} else {
			if err := m.update(l.Src.String(), l.Args[1]); err != nil {
				log.Printf("update failed for %q, %q: %s",
					l.Src.String(),
					l.Args[1:],
					err,
				)
			}
		}
	})

	log.Println("stat module initialised")
	return nil
}

func (m *StatMod) Reload () error {
	return nil
}

func (m *StatMod) Call (args ...string) error {
	return nil
}

func (m *StatMod) update (ident, message string) error {
	
	chars := len(message)
	words := Nwords(message)
	sentences := Nsentences(message)

	res, err := m.db.Exec(`
		UPDATE
			Stat
		SET
			chars = chars + ?,
			words = words + ?,
			sentences = sentences + ?,
			lines = lines + 1
		WHERE
			ident = ?`,
		chars,
		words,
		sentences,
		ident,
	)

	if err != nil {
		return err
	}

	if n, err := res.RowsAffected(); n < 1 {
		_, err = m.db.Exec(`
			INSERT INTO
				Stat (id, ident, chars, words, sentences, lines)
			VALUES (
				null, ?, ?, ?, ?, 1
			)`,
			ident, chars, words, sentences,
		)

		if err != nil {
			return err
		}
	}
	return nil
}

func Nwords (text string) int64 {
	s := bufio.NewScanner(strings.NewReader(text))
	s.Split(bufio.ScanWords)
	var count int64
	for s.Scan() {
		count++
	}
	return count
}

func Nsentences (text string) int64 {
	s := bufio.NewScanner(strings.NewReader(text))

	s.Split(func (data []byte, atEOF bool) (advance int, token []byte, err error) {
		start := 0
		width := 0
		for ; start < len(data); start += width {
			var r rune

			r, width = utf8.DecodeRune(data[start:])

			if !unicode.Is(unicode.STerm, r) {
				break
			}
		}

		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		for i := 0; i < len(data); i += width {
			var r rune

			r, width = utf8.DecodeRune(data[i:])

			if unicode.Is(unicode.STerm, r) {
				return i + width, data[start:i], nil
			}
		}

		if atEOF && len(data) > start {
			return len(data), data[start:], nil
		}

		return 0, nil, nil
	})

	var count int64

	for s.Scan() {
		count++
	}

	return count
}

type Stat struct {
	Ident string
	Chars, Words, Sentences, Lines int64
}

func (s Stat) String () string {
	return fmt.Sprintf("%s - %d chars, %d words, %d sentences, %d lines",
		s.Ident, s.Chars, s.Words, s.Sentences, s.Lines,
	)
}

func (m *StatMod) stat (ident string) (*Stat, error) {
	var s Stat

	err := m.db.QueryRow(`
		SELECT
			chars, words, sentences, lines
		FROM
			Stat
		WHERE
			ident = ?`,
		ident,
	).Scan(
		&s.Chars, &s.Words, &s.Sentences, &s.Lines,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	s.Ident = ident

	return &s, nil
}
