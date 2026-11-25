package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

const contentSchema = `
CREATE TABLE IF NOT EXISTS attachments(f TEXT PRIMARY KEY,t INTEGER,msg INTEGER,mime TEXT,data BLOB);
CREATE INDEX IF NOT EXISTS attachment_message ON attachments(t,msg);
CREATE TABLE IF NOT EXISTS agents(name TEXT, sub TEXT, hash TEXT UNIQUE,salt TEXT, PRIMARY KEY(name,sub));`

func (s *server) topicChange(w http.ResponseWriter, r *http.Request, t int, me string) {
	tx, e := s.db.Begin()
	checked(e)
	defer tx.Rollback()
	var note string
	if r.FormValue("a") == "rename" {
		title := taskText(r.FormValue("title"), 200)
		if title == "" {
			fail(w, 400, "нужно название")
			return
		}
		var old string
		checked(tx.QueryRow("SELECT title FROM topics WHERE id=?", t).Scan(&old))
		if old == title {
			respond(w, 200, obj{"ok": true})
			return
		}
		_, e = tx.Exec("UPDATE topics SET title=? WHERE id=?", title, t)
		note = fmt.Sprintf("%s переименовал(а) тему: «%s» → «%s»", me, old, title)
	} else {
		on := r.FormValue("on") != "0"
		var old bool
		checked(tx.QueryRow("SELECT coalesce((SELECT json_extract(body,'$.archived') FROM topic_metadata WHERE t=?),0)", t).Scan(&old))
		if on == old {
			respond(w, 200, obj{"ok": true})
			return
		}
		_, e = tx.Exec("INSERT INTO topic_metadata VALUES(?,json_object('archived',?)) ON CONFLICT(t) DO UPDATE SET body=json_set(body,'$.archived',?)", t, on, on)
		note = me + " вернул(а) тему из архива"
		if on {
			note = me + " убрал(а) тему в архив"
		}
	}
	checked(e)
	id := insertMessage(tx, t, obj{"sys": true, "who": me, "ai": false, "text": note, "ts": time.Now().UTC().Format(time.RFC3339)})
	checked(tx.Commit())
	s.signal(t)
	respond(w, 200, obj{"id": id})
}
func (s *server) editMessage(w http.ResponseWriter, r *http.Request, t int, me, sub string, ai bool) {
	on := number(r.FormValue("on"))
	tx, e := s.db.Begin()
	checked(e)
	defer tx.Rollback()
	var raw []byte
	e = tx.QueryRow("SELECT body FROM messages WHERE t=? AND "+mid+"=?"+ordinary, t, on).Scan(&raw)
	if errors.Is(e, sql.ErrNoRows) {
		fail(w, 404, "нет сообщения")
		return
	}
	checked(e)
	var m obj
	checked(json.Unmarshal(raw, &m))
	sig, _ := m["sub"].(string)
	if m["sys"] == true || m["gone"] == true || m["who"] != me || m["ai"] != ai || sig != sub {
		fail(w, 403, "можно менять только свои сообщения")
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	event := obj{"who": me, "sub": sub, "ai": ai, "priv": m["priv"] == true, "ts": now}
	if r.FormValue("a") == "edit" {
		text := taskText(r.FormValue("text"), 20000)
		if text == "" {
			fail(w, 400, "нужен текст")
			return
		}
		m["text"] = text
		m["edited"] = now
		event["edit"] = on
		event["text"] = text
		event["to"] = m["to"]
	} else {
		m = obj{"id": on, "who": me, "sub": sub, "ai": ai, "gone": true, "priv": event["priv"], "ts": m["ts"]}
		event["del"] = on
		_, e = tx.Exec("DELETE FROM attachments WHERE t=? AND msg=?", t, on)
		checked(e)
		_, e = tx.Exec("UPDATE messages SET body=json_remove(body,'$.text') WHERE t=? AND json_extract(body,'$.edit')=?", t, on)
		checked(e)
	}
	raw, e = json.Marshal(m)
	checked(e)
	_, e = tx.Exec("UPDATE messages SET body=? WHERE t=? AND "+mid+"=?", string(raw), t, on)
	checked(e)
	id := insertMessage(tx, t, event)
	checked(tx.Commit())
	s.signal(t)
	respond(w, 200, obj{"id": id})
}

var mediaTypes = map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".gif": "image/gif", ".webp": "image/webp", ".mp4": "video/mp4", ".webm": "video/webm", ".mov": "video/quicktime"}

func mediaType(ext string, data []byte) string {
	mime := mediaTypes[ext]
	if mime != "" && http.DetectContentType(data) == mime {
		if mime == "image/jpeg" || mime == "image/png" || mime == "image/gif" {
			if _, _, e := image.DecodeConfig(bytes.NewReader(data)); e != nil {
				return ""
			}
		}
		return mime
	}
	if ext == ".mov" && len(data) >= 12 && string(data[4:12]) == "ftypqt  " {
		return mime
	}
	return ""
}
func (s *server) savePost(w http.ResponseWriter, r *http.Request, t int, m obj) {
	tx, e := s.db.Begin()
	checked(e)
	defer tx.Rollback()
	if reply, ok := m["reply"].(int); ok {
		var raw []byte
		e := tx.QueryRow("SELECT body"+visible+" AND "+mid+"=?", t, m["who"], reply).Scan(&raw)
		if errors.Is(e, sql.ErrNoRows) {
			fail(w, 404, "нет такого сообщения")
			return
		}
		checked(e)
		var target struct {
			Who, Sub string
			AI       bool
			To       []string
		}
		checked(json.Unmarshal(raw, &target))
		if m["ai"] == true && !(target.AI && target.Who == m["who"] && target.Sub == stringField(m, "sub")) && len(target.To) > 0 && !slices.Contains(target.To, stringField(m, "from")) {
			fail(w, 409, "сообщение адресовано другому участнику")
			return
		}
	}
	if m["ai"] == true {
		var run int
		checked(tx.QueryRow("SELECT count(*) FROM (SELECT json_extract(body,'$.ai') ai FROM messages WHERE t=?"+ordinary+" AND NOT coalesce(json_extract(body,'$.sys'),0) AND NOT coalesce(json_extract(body,'$.priv'),0) ORDER BY "+mid+" DESC LIMIT 6) WHERE ai", t).Scan(&run))
		if run >= 6 {
			respond(w, 429, obj{"limit": true, "error": "Дождитесь ответа человека"})
			return
		}
	}
	if r.MultipartForm != nil {
		var files []obj
		for field, headers := range r.MultipartForm.File {
			if field != "file[]" && field != "file" {
				fail(w, 400, "неизвестное поле файла")
				return
			}
			for _, h := range headers {
				src, e := h.Open()
				checked(e)
				data, e := io.ReadAll(src)
				src.Close()
				checked(e)
				ext := strings.ToLower(filepath.Ext(h.Filename))
				mime := mediaType(ext, data)
				if mime == "" {
					fail(w, 400, "неподдерживаемый формат вложения")
					return
				}
				f := token("") + ext
				_, e = tx.Exec("INSERT INTO attachments VALUES(?,?,NULL,?,?)", f, t, mime, data)
				checked(e)
				files = append(files, obj{"f": f, "name": taskText(filepath.Base(h.Filename), 120), "size": len(data)})
			}
		}
		if len(files) > 0 {
			m["files"] = files
		}
	}
	id := insertMessage(tx, t, m)
	_, e = tx.Exec("UPDATE attachments SET msg=? WHERE t=? AND msg IS NULL", id, t)
	checked(e)
	checked(tx.Commit())
	if m["ai"] != true {
		s.activity(t, stringField(m, "who"), "human", 0)
	} else if stringField(m, "sub") == "" && m["priv"] != true {
		s.activity(t, stringField(m, "who"), "answer", id)
	}
	s.signal(t)
	respond(w, 200, obj{"id": id})
}
func (s *server) file(w http.ResponseWriter, r *http.Request, t int, me string) {
	var mime string
	var data []byte
	e := s.db.QueryRow("SELECT a.mime,a.data FROM attachments a JOIN messages m ON m.t=a.t AND coalesce(json_extract(m.body,'$.id'),m.id)=a.msg WHERE a.t=? AND a.f=? AND (coalesce(json_extract(m.body,'$.priv'),0)=0 OR json_extract(m.body,'$.who')=?)", t, r.FormValue("f"), me).Scan(&mime, &data)
	if errors.Is(e, sql.ErrNoRows) {
		fail(w, 404, "нет файла")
		return
	}
	checked(e)
	w.Header().Set("Content-Type", mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data))
}
