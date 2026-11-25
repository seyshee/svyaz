package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func readPHP(dir, name string, dst any) error {
	data, err := os.ReadFile(filepath.Join(dir, name+".json.php"))
	if err != nil {
		return err
	}
	header, body, ok := bytes.Cut(data, []byte("\n"))
	if !ok || strings.TrimSpace(string(header)) != "<?php exit; ?>" {
		return fmt.Errorf("%s: неизвестный заголовок", name)
	}
	if err = json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}
func optionalPHP(dir, name string, dst any) error {
	err := readPHP(dir, name, dst)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *server) importPHP(dir string) error {
	var signatures map[string][]string
	if err := legacyMap(dir, "ais", &signatures); err != nil {
		return err
	}
	var reports map[string]map[string]obj
	if err := legacyMap(dir, "use", &reports); err != nil {
		return err
	}
	var keys map[string]string
	var admins []string
	var secret string
	var topics []json.RawMessage
	if err := readPHP(dir, "keys", &keys); err != nil {
		return err
	}
	if len(keys) == 0 {
		return errors.New("пустой список пользователей")
	}
	if err := readPHP(dir, "admins", &admins); err != nil {
		return err
	}
	if err := optionalPHP(dir, "secret", &secret); err != nil {
		return err
	}
	if err := optionalPHP(dir, "topics", &topics); err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var used bool
	if err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM users UNION ALL SELECT 1 FROM topics UNION ALL SELECT 1 FROM messages)").Scan(&used); err != nil {
		return err
	}
	if used {
		return errors.New("для импорта нужна пустая база")
	}
	for hash, name := range keys {
		decoded, e := hex.DecodeString(hash)
		if e != nil || len(decoded) != 32 || name == "" {
			return errors.New("некорректная запись пользователя")
		}
		if _, err = tx.Exec("INSERT INTO users(name,h,admin) VALUES(?,?,?)", name, hash, slices.Contains(admins, name)); err != nil {
			return err
		}
		if secret != "" {
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(hash))
			for _, sub := range signatures[name] {
				mac.Reset()
				mac.Write([]byte(hash + "|" + sub))
				old := "ai-" + hex.EncodeToString(mac.Sum(nil))
				if _, err = tx.Exec("INSERT INTO agents VALUES(?,?,?,NULL)", name, sub, digest(old)); err != nil {
					return err
				}
			}
			mac.Reset()
			mac.Write([]byte(hash))
			ai := "ai-" + hex.EncodeToString(mac.Sum(nil))
			if _, err = tx.Exec("INSERT INTO legacy_ai VALUES(?,?)", digest(ai), name); err != nil {
				return err
			}
		}
	}
	for _, raw := range topics {
		var t struct {
			ID                 int
			Title, By, Created string
			Members            []string
			Gone               map[string]struct{ By, Why, Ts string }
		}
		if err = json.Unmarshal(raw, &t); err != nil {
			return err
		}
		if t.ID <= 0 {
			return errors.New("некорректный номер темы")
		}
		var messages []json.RawMessage
		if err = readPHP(dir, fmt.Sprintf("t%d", t.ID), &messages); err != nil {
			return err
		}
		legacyMembers := t.Members == nil
		if legacyMembers {
			t.Members = []string{t.By}
		}
		for _, body := range messages {
			var m struct {
				ID    int
				Who   string
				Files []struct{ F string }
			}
			if err = json.Unmarshal(body, &m); err != nil {
				return err
			}
			if m.ID <= 0 {
				return fmt.Errorf("тема %d: некорректный номер сообщения", t.ID)
			}
			for _, file := range m.Files {
				if !legacyFilename.MatchString(file.F) {
					return errors.New("некорректное имя вложения")
				}
				path := filepath.Join(dir, "f", file.F)
				info, e := os.Lstat(path)
				if e != nil {
					return e
				}
				if !info.Mode().IsRegular() || info.Size() > 64<<20 {
					return errors.New("некорректный файл вложения")
				}
				data, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				mime := mediaType(filepath.Ext(path), data)
				if mime == "" {
					return fmt.Errorf("неподдерживаемое вложение темы %d", t.ID)
				}
				if _, e = tx.Exec("INSERT INTO attachments VALUES(?,?,?,?,?)", file.F, t.ID, m.ID, mime, data); e != nil {
					return e
				}
			}
			if legacyMembers && m.Who != "" && !slices.Contains(t.Members, m.Who) {
				t.Members = append(t.Members, m.Who)
			}
			if _, err = tx.Exec("INSERT INTO messages(t,body) VALUES(?,?)", t.ID, string(body)); err != nil {
				return err
			}
		}
		var tasks []json.RawMessage
		if err = optionalPHP(dir, fmt.Sprintf("k%d", t.ID), &tasks); err != nil {
			return err
		}
		for _, body := range tasks {
			var task struct{ N int }
			if err = json.Unmarshal(body, &task); err != nil {
				return err
			}
			if task.N < 1 {
				return fmt.Errorf("тема %d: неверный номер задачи", t.ID)
			}
			if _, err = tx.Exec("INSERT INTO tasks VALUES(?,?,?)", t.ID, task.N, string(body)); err != nil {
				return err
			}
		}
		members, _ := json.Marshal(t.Members)
		if _, err = tx.Exec("INSERT INTO topics VALUES(?,?,?,?,?)", t.ID, t.Title, t.By, string(members), t.Created); err != nil {
			return err
		}
		for name, entry := range t.Gone {
			if slices.Contains(t.Members, name) {
				continue
			}
			if _, err = tx.Exec("INSERT INTO removed VALUES(?,?,?,?,?)", t.ID, name, entry.By, entry.Why, entry.Ts); err != nil {
				return err
			}
		}
		for n, report := range reports[fmt.Sprintf("t%d", t.ID)] {
			report["n"] = n
			data, e := json.Marshal(report)
			if e != nil {
				return e
			}
			if _, e = tx.Exec("INSERT INTO usage VALUES(?,?,?)", t.ID, n, string(data)); e != nil {
				return e
			}
		}
		if _, err = tx.Exec("INSERT INTO topic_metadata VALUES(?,?)", t.ID, string(raw)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

var legacyFilename = regexp.MustCompile(`^[0-9a-f]{32}\.[a-z0-9]{2,5}$`)

func legacyMap(dir, name string, dst any) error {
	var raw json.RawMessage
	if err := optionalPHP(dir, name, &raw); err != nil {
		return err
	}
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("[]")) {
		return nil
	}
	return json.Unmarshal(raw, dst)
}
