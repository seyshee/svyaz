package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func taskText(s string, limit int) string {
	r := []rune(strings.TrimSpace(s))
	return string(r[:min(len(r), limit)])
}

func (s *server) taskAPI(w http.ResponseWriter, r *http.Request, t int, me string, ai bool) {
	arg := r.FormValue
	if arg("a") == "tasks" {
		rows, err := s.db.Query("SELECT body FROM tasks WHERE t=? ORDER BY n", t)
		checked(err)
		defer rows.Close()
		list := []json.RawMessage{}
		for rows.Next() {
			var raw []byte
			checked(rows.Scan(&raw))
			list = append(list, json.RawMessage(raw))
		}
		checked(rows.Err())
		respond(w, 200, list)
		return
	}
	if !ai {
		fail(w, 403, "задачи ведёт ИИ темы")
		return
	}
	states := map[string]string{"now": "в работе", "todo": "не в работе", "done": "в архив"}
	state := arg("state")
	if state != "" && states[state] == "" {
		fail(w, 400, "state — now, todo или done")
		return
	}
	n := 0
	if arg("n") != "" {
		var err error
		n, err = strconv.Atoi(arg("n"))
		if err != nil || n < 0 {
			fail(w, 400, "некорректный номер задачи")
			return
		}
	}
	tx, err := s.db.Begin()
	checked(err)
	defer tx.Rollback()
	task := obj{"title": "", "state": "todo", "prio": 0, "info": "", "log": []any{}}
	changes := []string{}
	if n == 0 {
		checked(tx.QueryRow("SELECT coalesce(max(n),0)+1 FROM tasks WHERE t=?", t).Scan(&n))
		task["n"] = n
		changes = append(changes, "заведена")
	} else {
		var body []byte
		err = tx.QueryRow("SELECT body FROM tasks WHERE t=? AND n=?", t, n).Scan(&body)
		if errors.Is(err, sql.ErrNoRows) {
			fail(w, 404, "нет такой задачи")
			return
		}
		checked(err)
		checked(json.Unmarshal(body, &task))
	}
	if v := arg("title"); v != "" {
		task["title"] = taskText(v, 200)
		if len(changes) == 0 {
			changes = append(changes, "название")
		}
	}
	if state != "" {
		if task["state"] != state || len(changes) > 0 {
			changes = append(changes, states[state])
		}
		task["state"] = state
	}
	if v := arg("prio"); v != "" {
		task["prio"] = number(v)
		changes = append(changes, fmt.Sprintf("приоритет %d", task["prio"]))
	}
	if v := arg("info"); v != "" {
		task["info"] = taskText(v, 5000)
		changes = append(changes, "инфо обновлено")
	}
	if v := arg("note"); v != "" {
		changes = append(changes, taskText(v, 500))
	}
	now := time.Now().UTC().Format(time.RFC3339)
	entry := obj{"ts": now, "by": me, "what": strings.Join(changes, "; ")}
	if id := number(arg("msg")); id > 0 {
		entry["msg"] = id
	}
	history, _ := task["log"].([]any)
	task["log"] = append(history, entry)
	task["upd"] = now
	raw, err := json.Marshal(task)
	checked(err)
	_, err = tx.Exec("INSERT INTO tasks VALUES(?,?,?) ON CONFLICT(t,n) DO UPDATE SET body=excluded.body", t, n, string(raw))
	checked(err)
	checked(tx.Commit())
	respond(w, 200, obj{"n": n})
}
