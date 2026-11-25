package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed index.html kak.md
var assets embed.FS

type obj = map[string]any
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
func token(prefix string) string {
	var b [32]byte
	_, err := rand.Read(b[:])
	checked(err)
	return prefix + hex.EncodeToString(b[:])
}
func digest(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
func respond(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, msg string) { respond(w, status, obj{"error": msg}) }
func number(s string) int                                { n, _ := strconv.Atoi(s); return n }
func main() {
	addr := flag.String("listen", "127.0.0.1:8080", "HTTP listen address")
	dir := flag.String("data", "svyaz-data", "SQLite directory")
	source := flag.String("import-php", "", "import an offline PHP data directory and exit")
	flag.Parse()
	s := openServer(*dir)
	defer s.db.Close()
	if *source != "" {
		if err := s.importPHP(*source); err != nil {
			log.Fatal(err)
		}
		log.Print("Импорт завершён")
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api.php", s.api)
	mux.Handle("/", http.FileServer(http.FS(assets)))
	log.Fatal((&http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
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
func (s *server) api(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	defer func() {
		if e := recover(); e != nil {
			log.Print(e)
			fail(w, 500, "ошибка сервера")
		}
	}()
	r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			fail(w, 400, "ошибка формы")
			return
		}
	} else if err := r.ParseForm(); err != nil {
		fail(w, 400, "ошибка формы")
		return
	}
	arg := r.FormValue
	a := arg("a")
	writeAccess := a == "aikey" || a == "forget" || a == "kick" || a == "add"
	if writeAccess {
		s.access.Lock()
		defer s.access.Unlock()
	} else {
		s.access.RLock()
		defer s.access.RUnlock()
	}
	key := r.Header.Get("X-Key")
	if key == "" && a == "file" {
		if c, e := r.Cookie("svyaz_media"); e == nil {
			key = c.Value
		}
	}
	ai := strings.HasPrefix(key, "ai-")
	sub := ""
	var me string
	var admin bool
	err := s.db.QueryRow("SELECT name,admin FROM users WHERE h=? OR ai=? OR name=(SELECT name FROM legacy_ai WHERE hash=?)", digest(key), digest(key), digest(key)).Scan(&me, &admin)
	if errors.Is(err, sql.ErrNoRows) && ai {
		err = s.db.QueryRow("SELECT agents.name,sub FROM agents JOIN users ON users.name=agents.name WHERE hash=?", digest(key)).Scan(&me, &sub)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		panic(err)
	}
	if a != "setup" && err != nil {
		var empty bool
		checked(s.db.QueryRow("SELECT NOT EXISTS(SELECT 1 FROM users)").Scan(&empty))
		respond(w, 401, obj{"error": "нужен ключ", "setup": empty})
		return
	}
	agentLabel := aiLabel(me)
	if sub != "" {
		agentLabel += " · " + sub
	}
	admin = admin && !ai
	ok := func(v any) { respond(w, 200, v) }
	switch a {
	case "forget":
		s.forget(w, arg("who"), me, admin)
		return
	case "setup", "invite":
		if a == "invite" && !admin {
			fail(w, 403, "нужен администратор")
			return
		}
		n := strings.TrimSpace(arg("name"))
		if n == "" || utf8.RuneCountInString(n) > 40 || strings.ContainsAny(n, ",@<>") || strings.HasPrefix(strings.ToUpper(n), "ИИ") {
			fail(w, 400, "недопустимое имя")
			return
		}
		k := token("sv-")
		res, err := s.db.Exec("INSERT INTO users(name,h,admin) SELECT ?,?,? WHERE (? OR NOT EXISTS(SELECT 1 FROM users)) AND NOT EXISTS(SELECT 1 FROM retired WHERE name=?)", n, digest(k), a == "setup", a == "invite", n)
		if err == nil {
			count, e := res.RowsAffected()
			checked(e)
			if count == 0 {
				fail(w, 403, "уже настроено")
				return
			}
		}
		if err != nil {
			fail(w, 409, "имя занято или сервер уже настроен")
			return
		}
		ok(obj{"key": k, "name": n})
		return
	case "me":
		if !ai {
			http.SetCookie(w, &http.Cookie{Name: "svyaz_media", Value: key, Path: "/api.php", HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode})
		}
		ok(obj{"name": me, "ai": agentLabel, "admin": admin, "is_ai": ai, "subs": s.strings("SELECT sub FROM agents WHERE name=? ORDER BY sub", me), "admins": s.strings("SELECT name FROM users WHERE admin AND ? ORDER BY name", admin)})
		return
	case "aikey":
		if ai {
			fail(w, 403, "доступен основной ИИ человека")
			return
		}
		if sig := strings.TrimSpace(arg("sub")); sig != "" {
			if utf8.RuneCountInString(sig) > 20 || strings.ContainsAny(sig, ",@<>·") {
				fail(w, 400, "недопустимая подпись")
				return
			}
			if arg("off") == "1" {
				_, e := s.db.Exec("DELETE FROM agents WHERE name=? AND sub=?", me, sig)
				checked(e)
				s.wakeAll()
				ok(obj{"ok": true})
				return
			}
			salt := token("")
			var existing sql.NullString
			e := s.db.QueryRow("SELECT salt FROM agents WHERE name=? AND sub=?", me, sig).Scan(&existing)
			if e != nil && e != sql.ErrNoRows {
				checked(e)
			}
			if existing.Valid {
				salt = existing.String
			}
			k := "ai-" + digest("svyaz/sub/"+key+"/"+sig+"/"+salt)
			_, e = s.db.Exec("INSERT INTO agents VALUES(?,?,?,?) ON CONFLICT(name,sub) DO UPDATE SET hash=excluded.hash,salt=excluded.salt", me, sig, digest(k), salt)
			checked(e)
			s.wakeAll()
			ok(obj{"key": k, "ai": aiLabel(me) + " · " + sig})
			return
		}
		k := "ai-" + digest("svyaz/ai/"+key)
		_, err := s.db.Exec("UPDATE users SET ai=? WHERE name=?", digest(k), me)
		checked(err)
		ok(obj{"key": k})
		return
	case "people":
		rows, err := s.db.Query("SELECT name FROM users ORDER BY name")
		checked(err)
		defer rows.Close()
		names := []string{}
		for rows.Next() {
			var n string
			checked(rows.Scan(&n))
			names = append(names, n)
		}
		checked(rows.Err())
		ok(names)
		return
	case "new":
		if ai {
			fail(w, 403, "тему создаёт человек")
			return
		}
		title := strings.TrimSpace(arg("title"))
		if title == "" {
			fail(w, 400, "нужен title")
			return
		}
		members := []string{me}
		for _, n := range strings.Split(arg("members"), ",") {
			n = strings.TrimSpace(n)
			if n == "" || slices.Contains(members, n) {
				continue
			}
			var exists bool
			checked(s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE name=?)", n).Scan(&exists))
			if !exists {
				fail(w, 400, "нет такого участника")
				return
			}
			members = append(members, n)
		}
		raw, _ := json.Marshal(members)
		res, err := s.db.Exec("INSERT INTO topics(title,creator,members,created) VALUES(?,?,?,?)", title, me, string(raw), time.Now().UTC().Format(time.RFC3339))
		checked(err)
		id, err := res.LastInsertId()
		checked(err)
		ok(obj{"id": id})
		return
	case "gone":
		ok(s.removals(me, 0))
		return
	case "holiday":
		day := today()
		next := day.AddDate(0, 0, 1)
		tomorrow := holiday(next)
		if tomorrow["from"] != next.Format(time.DateOnly) {
			tomorrow = nil
		}
		ok(obj{"today": holiday(day), "tomorrow": tomorrow})
		return
	}
	t := number(arg("t"))
	rows, err := s.db.Query(`SELECT p.id,title,creator,members,created,
 coalesce((SELECT json_extract(body,'$.archived') FROM topic_metadata WHERE t=p.id),0),
 (SELECT count(*) FROM messages WHERE t=p.id AND (coalesce(json_extract(body,'$.priv'),0)=0 OR json_extract(body,'$.who')=?)),
 coalesce((SELECT json_extract(body,'$.ts') FROM messages WHERE t=p.id AND (coalesce(json_extract(body,'$.priv'),0)=0 OR json_extract(body,'$.who')=?) ORDER BY coalesce(json_extract(body,'$.id'),id) DESC LIMIT 1),created)
 FROM topics p WHERE (?='topics' OR p.id=?) AND EXISTS(SELECT 1 FROM json_each(members) WHERE value=?)`, me, me, a, t, me)
	checked(err)
	defer rows.Close()
	topics := []obj{}
	for rows.Next() {
		var id int
		var title, by, raw, created string
		var archived bool
		var count int
		var last string
		checked(rows.Scan(&id, &title, &by, &raw, &created, &archived, &count, &last))
		topics = append(topics, obj{"archived": archived, "id": id, "title": title, "by": by, "members": json.RawMessage(raw), "created": created, "count": count, "last": last, "calls": []any{}, "ais": obj{}})
	}
	checked(rows.Err())
	rows.Close()
	if a == "topics" {
		for _, topic := range topics {
			id := topic["id"].(int)
			topic["ais"] = s.topicAgents(id)
			topic["calls"] = s.calls(id, me)
		}
		ok(topics)
		return
	}
	if len(topics) == 0 {
		if gone := s.removals(me, t); t != 0 && len(gone) > 0 {
			respond(w, 410, obj{"gone": true, "error": fmt.Sprint("Доступ к теме отозван. ", gone[0]["why"])})
			return
		}
		fail(w, 404, "нет такой темы")
		return
	}
	switch a {
	case "add", "kick":
		s.membership(w, r, t, me, ai, admin, topics[0])
		return
	case "edit", "del":
		s.editMessage(w, r, t, me, sub, ai)
	case "rename", "archive":
		s.topicChange(w, r, t, me)
	case "usage":
		if !ai {
			fail(w, 403, "расход шлёт ИИ")
			return
		}
		s.reportUsage(w, t, agentLabel, arg("u"))
	case "typing":
		if ai {
			fail(w, 403, "печатает человек")
			return
		}
		s.activity(t, me, "typing", 0)
		ok(obj{"ok": true})
	case "file":
		s.file(w, r, t, me)
	case "post":
		text := strings.TrimSpace(arg("text"))
		if (text == "" && arg("card") != "1" && (r.MultipartForm == nil || len(r.MultipartForm.File) == 0)) || utf8.RuneCountInString(text) > 20000 {
			fail(w, 400, "доступен текст до 20000 символов")
			return
		}
		from := me
		if ai {
			from = agentLabel
		}
		m := obj{"who": me, "from": from, "ai": ai, "text": text, "ts": time.Now().UTC().Format(time.RFC3339), "priv": arg("priv") == "1"}
		if sub != "" {
			m["sub"] = sub
		}
		if arg("ask") == "1" {
			m["ask"] = true
		}
		if arg("card") == "1" {
			h := holiday(today())
			if h == nil {
				fail(w, 400, "сегодня нет праздника")
				return
			}
			m["card"] = obj{"name": h["name"], "hue": h["hue"]}
		}
		if arg("priv") != "1" {
			to, e := s.recipients(topics[0]["members"].(json.RawMessage), text, arg("to"))
			if e != nil {
				fail(w, 400, e.Error())
				return
			}
			if len(to) > 0 {
				m["to"] = to
			}
		}
		if reply := number(arg("reply")); reply > 0 {
			m["reply"] = reply
			if quote := taskText(arg("quote"), 1000); quote != "" {
				m["quote"] = quote
			}
		}
		s.savePost(w, r, t, m)
	case "tasks", "task":
		s.taskAPI(w, r, t, agentLabel, ai)
	case "react":
		meaning, exists := reactionMeanings[arg("r")]
		if !exists {
			fail(w, 400, "неизвестная реакция")
			return
		}
		on := number(arg("on"))
		var owner, targetSub string
		var private, targetAI bool
		err := s.db.QueryRow(`SELECT json_extract(body,'$.who'),coalesce(json_extract(body,'$.priv'),0),coalesce(json_extract(body,'$.ai'),0),coalesce(json_extract(body,'$.sub'),'') FROM messages WHERE t=? AND coalesce(json_extract(body,'$.id'),id)=? AND coalesce(json_extract(body,'$.gone'),0)=0 AND json_extract(body,'$.react') IS NULL AND json_extract(body,'$.edit') IS NULL AND json_extract(body,'$.del') IS NULL`, t, on).Scan(&owner, &private, &targetAI, &targetSub)
		if errors.Is(err, sql.ErrNoRows) || private && owner != me {
			fail(w, 404, "нет сообщения")
			return
		}
		checked(err)
		off := arg("off") == "1"
		if off {
			meaning += " (снято)"
		}
		m := obj{"who": me, "ai": ai, "sub": sub, "on": on, "off": off, "priv": private, "react": arg("r"), "meaning": meaning, "ts": time.Now().UTC().Format(time.RFC3339)}
		if targetAI {
			m["notify_ai"] = owner
			m["notify_sub"] = targetSub
		}
		ok(obj{"id": s.appendMessage(t, m)})
	case "read", "wait":
		if a == "wait" && ai && sub == "" {
			s.activity(t, me, "wait", 0)
		}
		seconds := number(arg("wait"))
		if a == "wait" && arg("wait") == "" {
			seconds = 25
		}
		timer := time.NewTimer(time.Duration(min(max(seconds, 0), 60)) * time.Second)
		defer timer.Stop()
		for {
			s.mu.Lock()
			ch := s.changed[t]
			if ch == nil {
				ch = make(chan struct{})
				s.changed[t] = ch
			}
			s.mu.Unlock()
			if sub != "" {
				var valid bool
				checked(s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM agents WHERE hash=?)", digest(key)).Scan(&valid))
				if !valid {
					fail(w, 401, "ключ отозван")
					return
				}
			}
			var allowed bool
			checked(s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM topics,json_each(members) WHERE topics.id=? AND value=?)", t, me).Scan(&allowed))
			if !allowed {
				respond(w, 410, obj{"gone": true, "error": "Доступ к теме отозван"})
				return
			}
			rows, skip, j, err := s.history(t, me, arg)
			checked(err)
			messages := []obj{}
			hit := false
			for rows.Next() {
				var id int
				var raw []byte
				checked(rows.Scan(&id, &raw))
				var m obj
				checked(json.Unmarshal(raw, &m))
				m["id"] = id
				messages = append(messages, m)
				if m["react"] != nil {
					hit = hit || (m["notify_ai"] == me && stringField(m, "notify_sub") == sub && !(m["ai"] == true && m["who"] == me && stringField(m, "sub") == sub))
				} else if m["sys"] != true && m["del"] == nil && (m["ai"] != true || m["who"] != me || stringField(m, "sub") != sub) {
					if tos, yes := m["to"].([]any); !yes || slices.Contains(tos, any(agentLabel)) {
						hit = true
					}
				}
			}
			checked(rows.Err())
			rows.Close()
			result := obj{"messages": messages, "skip": skip, "j": j, "topic": topics[0], "me": me, "my_ai": agentLabel, "holiday": holiday(today())}
			if a == "read" {
				for k, v := range s.live(t, me) {
					result[k] = v
				}
			}
			if a == "wait" && hit && ai && sub == "" {
				on := 0
				for _, m := range messages {
					if m["react"] == nil && m["edit"] == nil && m["del"] == nil && m["sys"] != true && m["priv"] != true && !(m["ai"] == true && m["who"] == me && stringField(m, "sub") == "") {
						on = m["id"].(int)
					}
				}
				s.activity(t, me, "hit", on)
			}
			if (a == "wait" && hit) || (a == "read" && (len(messages) > 0 || arg("lv") != "" && arg("lv") != result["lv"])) {
				ok(result)
				return
			}
			s.access.RUnlock()
			select {
			case <-ch:
				s.access.RLock()
				continue
			case <-r.Context().Done():
				s.access.RLock()
				return
			case <-timer.C:
				s.access.RLock()
				if a == "wait" {
					w.WriteHeader(204)
				} else {
					for k, v := range s.live(t, me) {
						result[k] = v
					}
					ok(result)
				}
				return
			}
		}
	default:
		fail(w, 501, "действие пока не поддерживается Go-сервером")
	}
}
func aiLabel(name string) string {
	words := strings.Fields(name)
	for i, w := range words {
		r := []rune(w)
		end := len(r) - 1
		surn := false
		for _, s := range []string{"ова", "ева", "ёва", "ина", "ына"} {
			surn = surn || strings.HasSuffix(w, s)
		}
		switch {
		case len(r) > 4 && strings.HasSuffix(w, "ая"):
			words[i] = string(r[:end-1]) + "ой"
		case len(r) > 4 && surn && !slices.Contains(strings.Fields("Ирина Марина Алина Полина Карина Галина Регина Кристина Валентина Екатерина Ангелина Арина Зарина Альбина Каролина Эвелина Мальвина Сабина Устина Акулина"), w):
			words[i] = string(r[:end]) + "ой"
		case r[end] == 'я' || r[end] == 'а' && end > 0 && strings.ContainsRune("гкхжчшщ", r[end-1]):
			words[i] = string(r[:end]) + "и"
		case r[end] == 'а':
			words[i] = string(r[:end]) + "ы"
		case strings.ContainsRune("йь", r[end]):
			words[i] = string(r[:end]) + "я"
		case !strings.ContainsRune("еёиоуыэю", r[end]):
			words[i] = w + "а"
		}
	}
	return "ИИ " + strings.Join(words, " ")
}

var reactionMeanings = map[string]string{
	"agree":    "согласен",
	"disagree": "не согласен",
	"seen":     "видел, принял",
	"question": "непонятно, поясни",
	"done":     "сделано",
	"smile":    "смешно, приятно",
	"kiss":     "поцелуйчик, спасибо",
	"thanks":   "спасибо, отвечать не нужно",
	"fire":     "огонь, очень нравится — так и дальше",
	"deal":     "договорились: решение принято, делай по последнему предложению",
	"look":     "иду смотреть: человек пошёл проверять, жди его слова и не дёргай",
	"cool":     "прикольно: забавно и симпатично, отвечать не нужно",
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
