package main

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const liveSchema = `CREATE TABLE IF NOT EXISTS activity(t INTEGER,name TEXT,typing INTEGER DEFAULT 0,waiting INTEGER DEFAULT 0,thinking INTEGER DEFAULT 0,on_msg INTEGER DEFAULT 0,answer INTEGER DEFAULT 0,PRIMARY KEY(t,name));
CREATE TABLE IF NOT EXISTS usage(t INTEGER,name TEXT,body TEXT,PRIMARY KEY(t,name));`

func (s *server) activity(t int, name, event string, id int) {
	now := time.Now().Unix()
	var query string
	switch event {
	case "typing":
		query = "INSERT INTO activity(t,name,typing) VALUES(?,?,?) ON CONFLICT(t,name) DO UPDATE SET typing=excluded.typing"
	case "human":
		_, e := s.db.Exec("UPDATE activity SET typing=0 WHERE t=? AND name=? AND typing<>0", t, name)
		checked(e)
		return
	case "wait":
		query = "INSERT INTO activity(t,name,waiting) VALUES(?,?,?) ON CONFLICT(t,name) DO UPDATE SET waiting=excluded.waiting,thinking=0,on_msg=CASE WHEN answer=0 THEN 0 ELSE on_msg END"
	case "hit":
		_, e := s.db.Exec("UPDATE activity SET thinking=?,waiting=?,on_msg=CASE WHEN ?>0 THEN ? ELSE on_msg END,answer=CASE WHEN ?>0 THEN 0 ELSE answer END WHERE t=? AND name=?", now, now, id, id, id, t, name)
		checked(e)
		s.signal(t)
		return
	case "answer":
		_, e := s.db.Exec("UPDATE activity SET answer=?,thinking=? WHERE t=? AND name=? AND on_msg>0 AND answer=0", id, now, t, name)
		checked(e)
		return
	}
	_, e := s.db.Exec(query, t, name, now)
	checked(e)
	s.signal(t)
}
func (s *server) live(t int, me string) obj {
	states := obj{}
	typing := []string{}
	work := []obj{}
	usage := []json.RawMessage{}
	now := time.Now().Unix()
	rows, e := s.db.Query("SELECT name,typing,waiting,thinking,on_msg,answer FROM activity JOIN topics ON id=t WHERE t=? AND name IN(SELECT value FROM json_each(members)) ORDER BY name", t)
	checked(e)
	for rows.Next() {
		var n string
		var typ, wait, think int64
		var on, ans int
		checked(rows.Scan(&n, &typ, &wait, &think, &on, &ans))
		if typ > now-6 && n != me {
			typing = append(typing, n)
		}
		if wait > 0 {
			state := "off"
			if think > now-900 {
				state = "think"
			} else if wait > now-60 {
				state = "on"
			}
			states[n] = state
		}
		if on > 0 && (ans > 0 || think > now-900) {
			work = append(work, obj{"n": n, "on": on, "ans": ans})
		}
	}
	checked(rows.Err())
	rows.Close()
	rows, e = s.db.Query("SELECT body FROM usage WHERE t=? ORDER BY name", t)
	checked(e)
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		checked(rows.Scan(&raw))
		usage = append(usage, json.RawMessage(raw))
	}
	checked(rows.Err())
	v := obj{"ai_state": states, "typing": typing, "work": work, "usage": usage}
	b, e := json.Marshal(v)
	checked(e)
	v["lv"] = digest(string(b))[:12]
	return v
}

// Rates match the legacy usage report, in USD per million tokens.
var rates = map[string][3]float64{"claude-opus-5-5": {4, 20, .2}, "claude-opus-5": {5, 25, .5}, "claude-opus-4": {5, 25, .5}, "claude-fable-5-1": {10, 50, .25}, "claude-fable-5": {10, 50, 1}, "claude-sonnet-5": {2, 10, .2}, "claude-sonnet-4": {3, 15, .3}, "claude-haiku-4": {1, 5, .1}}

func (s *server) reportUsage(w http.ResponseWriter, t int, name, input string) {
	var tokens int64
	var cost float64
	since := ""
	for _, line := range strings.Split(input, "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && f[0] == "since" {
			since = taskText(f[1], 30)
			continue
		}
		if len(f) != 6 {
			continue
		}
		var nums [5]float64
		for i, v := range f[1:] {
			n, e := strconv.ParseUint(v, 10, 50)
			if e != nil {
				fail(w, 400, "неверное число токенов")
				return
			}
			nums[i] = float64(n)
			tokens += int64(n)
		}
		var rate [3]float64
		best := 0
		for model, r := range rates {
			if strings.HasPrefix(f[0], model) && len(model) > best {
				rate = r
				best = len(model)
			}
		}
		cost += (nums[0]*rate[0] + nums[1]*rate[2] + nums[2]*1.25*rate[0] + nums[3]*2*rate[0] + nums[4]*rate[1]) / 1e6
	}
	tx, e := s.db.Begin()
	checked(e)
	defer tx.Rollback()
	var raw []byte
	var old struct {
		Since, Ts string
		Tok       int64
		Usd       float64
		Hist      []obj
	}
	e = tx.QueryRow("SELECT body FROM usage WHERE t=? AND name=?", t, name).Scan(&raw)
	if e != sql.ErrNoRows {
		checked(e)
		checked(json.Unmarshal(raw, &old))
	}
	hist := []obj{}
	if old.Since != "" && old.Since != since {
		hist = append(hist, obj{"since": old.Since, "till": old.Ts, "tok": old.Tok, "usd": old.Usd})
	}
	for _, entry := range old.Hist {
		if entry["since"] != since && len(hist) < 10 {
			hist = append(hist, entry)
		}
	}
	v := obj{"n": name, "since": since, "tok": tokens, "usd": math.Round(cost*100) / 100, "hist": hist, "ts": time.Now().UTC().Format(time.RFC3339)}
	raw, e = json.Marshal(v)
	checked(e)
	_, e = tx.Exec("INSERT OR REPLACE INTO usage VALUES(?,?,?)", t, name, string(raw))
	checked(e)
	checked(tx.Commit())
	s.signal(t)
	respond(w, 200, v)
}
