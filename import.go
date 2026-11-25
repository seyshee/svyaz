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
				ID  int
				Who string
			}
			if err = json.Unmarshal(body, &m); err != nil {
				return err
			}
			if m.ID <= 0 {
				return fmt.Errorf("тема %d: некорректный номер сообщения", t.ID)
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
		if _, err = tx.Exec("INSERT INTO topic_metadata VALUES(?,?)", t.ID, string(raw)); err != nil {
			return err
		}
	}
	return tx.Commit()
}
