package main

import (
	czql "github.com/cznic/ql"
	"upper.io/db"
	"upper.io/db/ql"
)

var schema = `
BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS prefixes (
	prefix string NOT NULL,
);

CREATE TABLE IF NOT EXISTS suffixes (
	suffix string NOT NULL,
	prefixID int64,
	count int64,	-- number of occurances
);

COMMIT;
`

func initdb(path string) error {
	db, err := czql.OpenFile(path, &czql.Options{CanCreate: true})
	if err != nil {
		return err
	}

	defer db.Close()

	_, _, err = db.Run(czql.NewRWCtx(), schema, nil)
	if err != nil {
		return err
	}

	return nil
}

type Prefix struct {
	ID     int64  `db:"id,omitempty"`
	Prefix string `db:"prefix"`
}

func (p *Prefix) SetID(id int64) error {
	p.ID = id
	return nil
}

type Suffix struct {
	ID       int64  `db:"id,omitempty"`
	Suffix   string `db:"suffix"`
	PrefixID int64  `db:"prefixID"`
	Count    int64  `db:"count"`
}

func (s *Suffix) SetID(id int64) error {
	s.ID = id
	return nil
}

type MarkovDB interface {
	InsertPrefix(prefix string) (int64, error)
	InsertSuffix(prefixID int64, suffix string) (int64, error)
	Suffixes(prefix string) ([]Suffix, error)
	Close() error
}

type markovDB struct {
	dbConn db.Database
}

func NewMarkovDB(conn string) (MarkovDB, error) {
	if err := initdb(conn); err != nil {
		return nil, err
	}

	dbURL := ql.ConnectionURL{
		Database: conn, // Path to a QL database file.
	}

	err := initdb(dbURL.Database)
	if err != nil {
		return nil, err
	}

	dbConn, err := db.Open(ql.Adapter, dbURL)
	if err != nil {
		return nil, err
	}

	return &markovDB{dbConn}, nil
}

func (m *markovDB) InsertPrefix(prefix string) (int64, error) {
	prefixes, err := m.dbConn.Collection("prefixes")
	if err != nil {
		return 0, err
	}

	pre := Prefix{Prefix: prefix}

	_, err = prefixes.Append(&pre)
	if err != nil {
		return 0, err
	}

	q := prefixes.Find(db.Cond{"prefix": prefix}).Select("id() as id")
	err = q.One(&pre)
	return pre.ID, err
}

func (m *markovDB) InsertSuffix(prefixID int64, suffix string) (int64, error) {
	suffixes, err := m.dbConn.Collection("suffixes")
	if err != nil {
		return 0, err
	}

	suf := Suffix{
		PrefixID: prefixID,
		Suffix:   suffix,
		Count:    1,
	}

	_, err = suffixes.Append(&suf)
	if err != nil {
		return 0, err
	}

	q := suffixes.Find(db.Cond{"suffix": suffix}).Select("id() as id")
	err = q.One(&suf)
	return suf.ID, err
}

func (m *markovDB) Suffixes(prefix string) ([]Suffix, error) {
	prefixes, err := m.dbConn.Collection("prefixes")
	if err != nil {
		return nil, err
	}

	var pre Prefix
	res := prefixes.Find(db.Cond{"prefix": prefix})
	err = res.One(&pre)
	if err != nil {
		return nil, err
	}

	suffixes, err := m.dbConn.Collection("suffixes")
	if err != nil {
		return nil, err
	}

	var suf []Suffix
	res = suffixes.Find(db.Cond{"suffixID": pre.ID}).Select("id() as id", "suffix", "prefixID", "count")
	err = res.All(&suf)
	if err != nil {
		return nil, err
	}

	return suf, nil
}

func (m *markovDB) Close() error {
	return m.dbConn.Close()
}
