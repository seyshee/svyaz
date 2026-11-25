package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"
)

func token(prefix string) string {
	var b [32]byte
	_, err := rand.Read(b[:])
	checked(err)
	return prefix + hex.EncodeToString(b[:])
}
func digest(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }

type identity struct {
	key, name, sub, label string
	ai, admin             bool
}

func (s *server) authenticate(w http.ResponseWriter, r *http.Request, a string) (identity, bool) {
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
		return identity{}, false
	}
	agentLabel := aiLabel(me)
	if sub != "" {
		agentLabel += " · " + sub
	}
	admin = admin && !ai
	return identity{key: key, name: me, sub: sub, label: agentLabel, ai: ai, admin: admin}, true
}
func (s *server) account(w http.ResponseWriter, r *http.Request, user identity) {
	arg := r.FormValue
	a := arg("a")
	me, key, admin, ai, agentLabel := user.name, user.key, user.admin, user.ai, user.label
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
	}
}
