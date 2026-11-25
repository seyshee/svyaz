package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"
)

func (s *server) removals(name string, t int) []obj {
	rows, e := s.db.Query("SELECT topics.id,title,by_name,reason,ts FROM removed JOIN topics ON topics.id=t WHERE name=? AND (?=0 OR t=?)", name, t, t)
	checked(e)
	defer rows.Close()
	result := []obj{}
	for rows.Next() {
		var id int
		var title, by, why, ts string
		checked(rows.Scan(&id, &title, &by, &why, &ts))
		result = append(result, obj{"id": id, "title": title, "by": by, "why": why, "ts": ts})
	}
	checked(rows.Err())
	return result
}
func (s *server) wakeAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, ch := range s.changed {
		close(ch)
		delete(s.changed, id)
	}
}
func (s *server) forget(w http.ResponseWriter, name, me string, admin bool) {
	name = strings.TrimSpace(name)
	if !admin || name == me {
		fail(w, 403, "удаление запрещено")
		return
	}
	tx, e := s.db.Begin()
	checked(e)
	defer tx.Rollback()
	res, e := tx.Exec("DELETE FROM users WHERE name=? AND admin=0", name)
	checked(e)
	n, e := res.RowsAffected()
	checked(e)
	if n == 0 {
		fail(w, 403, "пользователя нельзя удалить")
		return
	}
	_, e = tx.Exec("INSERT INTO retired VALUES(?)", name)
	checked(e)
	_, e = tx.Exec("DELETE FROM legacy_ai WHERE name=?", name)
	checked(e)
	_, e = tx.Exec("UPDATE topics SET members=(SELECT json_group_array(value) FROM json_each(members) WHERE value<>?) WHERE EXISTS(SELECT 1 FROM json_each(members) WHERE value=?)", name, name)
	checked(e)
	checked(tx.Commit())
	s.wakeAll()
	respond(w, 200, obj{"ok": true})
}
func (s *server) membership(w http.ResponseWriter, r *http.Request, t int, me string, ai, admin bool, topic obj) {
	add := r.FormValue("a") == "add"
	if ai || !add && !admin && topic["by"] != me {
		fail(w, 403, "нет прав менять участников")
		return
	}
	var members []string
	checked(json.Unmarshal(topic["members"].(json.RawMessage), &members))
	name := strings.TrimSpace(r.FormValue("who"))
	tx, e := s.db.Begin()
	checked(e)
	defer tx.Rollback()
	why := taskText(r.FormValue("why"), 200)
	ts := time.Now().UTC().Format(time.RFC3339)
	var note string
	if add {
		var exists bool
		checked(tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE name=?)", name).Scan(&exists))
		if !exists {
			fail(w, 400, "нет пользователя")
			return
		}
		if slices.Contains(members, name) {
			respond(w, 200, obj{"ok": true})
			return
		}
		members = append(members, name)
		_, e = tx.Exec("DELETE FROM removed WHERE t=? AND name=?", t, name)
		checked(e)
		note = me + " позвал(а) в тему: " + name
	} else {
		if name == me {
			fail(w, 400, "себя убрать нельзя")
			return
		}
		gone := []string{}
		keep := []string{}
		for _, n := range members {
			if n != me && (r.FormValue("all") == "1" || n == name) {
				gone = append(gone, n)
				_, e = tx.Exec("INSERT OR REPLACE INTO removed VALUES(?,?,?,?,?)", t, n, me, why, ts)
				checked(e)
			} else {
				keep = append(keep, n)
			}
		}
		if len(gone) == 0 {
			respond(w, 200, obj{"ok": true})
			return
		}
		members = keep
		note = me + " убрал(а) из темы: " + strings.Join(gone, ", ")
		if why != "" {
			note += ". Причина: " + why
		}
	}
	raw, e := json.Marshal(members)
	checked(e)
	_, e = tx.Exec("UPDATE topics SET members=? WHERE id=?", string(raw), t)
	checked(e)
	id := insertMessage(tx, t, obj{"sys": true, "who": me, "ai": false, "text": note, "ts": ts})
	checked(tx.Commit())
	s.wakeAll()
	respond(w, 200, obj{"id": id})
}

type messageDB interface{ QueryRow(string, ...any) *sql.Row }
