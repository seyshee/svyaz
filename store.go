package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"sync"
)

type server struct {
	access  sync.RWMutex
	db      *sql.DB
	mu      sync.Mutex
	changed map[int]chan struct{}
}

func checked(err error) {
	if err != nil {
		panic(err)
	}
}
func openServer(dir string) *server {
	checked(os.MkdirAll(dir, 0700))
	db, err := sql.Open("sqlite3", filepath.Join(dir, "chat.sqlite")+"?_txlock=immediate&_journal_mode=WAL&_busy_timeout=5000")
	checked(err)
	db.SetMaxOpenConns(4)
	_, err = db.Exec(liveSchema + contentSchema + `CREATE TABLE IF NOT EXISTS users(name TEXT PRIMARY KEY,h TEXT UNIQUE, ai TEXT UNIQUE, admin BOOLEAN);
 CREATE TABLE IF NOT EXISTS removed(t INTEGER,name TEXT,by_name TEXT,reason TEXT,ts TEXT,PRIMARY KEY(t,name)); CREATE TABLE IF NOT EXISTS retired(name TEXT PRIMARY KEY); CREATE TABLE IF NOT EXISTS tasks(t INTEGER,n INTEGER,body TEXT,PRIMARY KEY(t,n)); DROP INDEX IF EXISTS founder; CREATE TABLE IF NOT EXISTS legacy_ai(hash TEXT PRIMARY KEY,name TEXT); CREATE TABLE IF NOT EXISTS topic_metadata(t INTEGER PRIMARY KEY,body TEXT);
 CREATE TABLE IF NOT EXISTS topics(id INTEGER PRIMARY KEY,title TEXT,creator TEXT,members TEXT,created TEXT);
 CREATE TABLE IF NOT EXISTS messages(id INTEGER PRIMARY KEY,t INTEGER,body TEXT);
 CREATE INDEX IF NOT EXISTS history ON messages(t,id); CREATE UNIQUE INDEX IF NOT EXISTS local_ids ON messages(t,coalesce(json_extract(body,'$.id'),id))`)
	checked(err)
	return &server{db: db, changed: make(map[int]chan struct{})}
}
func (s *server) appendMessage(t int, m obj) int {
	id := insertMessage(s.db, t, m)
	s.signal(t)
	return id
}
func (s *server) signal(t int) {
	s.mu.Lock()
	if ch := s.changed[t]; ch != nil {
		close(ch)
		delete(s.changed, t)
	}
	s.mu.Unlock()

}

func insertMessage(db messageDB, t int, m obj) int {
	raw, err := json.Marshal(m)
	checked(err)
	var id int
	err = db.QueryRow(`INSERT INTO messages(t,body) SELECT ?,json_set(?, '$.id',1+coalesce(max(coalesce(json_extract(body,'$.id'),id)),0)) FROM messages WHERE t=? RETURNING json_extract(body,'$.id')`, t, string(raw), t).Scan(&id)
	checked(err)
	return id
}

func stringField(m obj, k string) string { v, _ := m[k].(string); return v }
func (s *server) strings(q string, args ...any) []string {
	rows, e := s.db.Query(q, args...)
	checked(e)
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var v string
		checked(rows.Scan(&v))
		out = append(out, v)
	}
	checked(rows.Err())
	return out
}
