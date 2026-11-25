package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
)

func (s *server) topicAgents(t int) map[string][]string {
	rows, e := s.db.Query(`SELECT DISTINCT a.name,a.sub FROM agents a JOIN topics p ON p.id=? AND a.name IN(SELECT value FROM json_each(p.members)) JOIN messages m ON m.t=p.id AND json_extract(m.body,'$.who')=a.name AND json_extract(m.body,'$.sub')=a.sub ORDER BY a.name,a.sub`, t)
	checked(e)
	defer rows.Close()
	result := map[string][]string{}
	for rows.Next() {
		var n, sub string
		checked(rows.Scan(&n, &sub))
		result[n] = append(result[n], sub)
	}
	checked(rows.Err())
	return result
}
func (s *server) recipients(raw json.RawMessage, text, to string) ([]string, error) {
	var members []string
	checked(json.Unmarshal(raw, &members))
	var allowed []string
	for _, n := range members {
		allowed = append(allowed, n, aiLabel(n))
		for _, sub := range s.strings("SELECT sub FROM agents WHERE name=?", n) {
			allowed = append(allowed, aiLabel(n)+" · "+sub)
		}
	}
	var result []string
	for _, v := range strings.Split(to, ",") {
		n := strings.TrimSpace(v)
		if n == "" {
			continue
		}
		if !slices.Contains(allowed, n) {
			return nil, fmt.Errorf("в теме нет «%s»", n)
		}
		if !slices.Contains(result, n) {
			result = append(result, n)
		}
	}
	sort.SliceStable(allowed, func(i, j int) bool { return len(allowed[i]) > len(allowed[j]) })
	rest := strings.ToLower(text)
	for _, n := range allowed {
		mention := "@" + strings.ToLower(n)
		if strings.Contains(rest, mention) {
			if !slices.Contains(result, n) {
				result = append(result, n)
			}
			rest = strings.ReplaceAll(rest, mention, "")
		}
	}
	return result, nil
}
func (s *server) calls(t int, me string) []int {
	// Only questions since the person's latest ordinary message can be pending.
	rows, e := s.db.Query("SELECT "+mid+",body"+visible+" AND "+mid+">coalesce((SELECT max("+mid+") FROM messages WHERE t=? AND json_extract(body,'$.who')=? AND NOT json_extract(body,'$.ai') AND NOT coalesce(json_extract(body,'$.sys'),0)"+ordinary+"),0) AND (EXISTS(SELECT 1 FROM json_each(body,'$.to') WHERE value=?) OR (json_extract(body,'$.who')=? AND NOT json_extract(body,'$.ai') AND json_extract(body,'$.react') IS NOT NULL)) ORDER BY "+mid, t, me, t, me, me, me)
	checked(e)
	defer rows.Close()
	out := []int{}
	for rows.Next() {
		var id int
		var raw []byte
		checked(rows.Scan(&id, &raw))
		var m struct {
			Text, React   string
			Ask           bool
			On, Edit, Del int
		}
		checked(json.Unmarshal(raw, &m))
		if m.React != "" {
			out = slices.DeleteFunc(out, func(id int) bool { return id == m.On })
			continue
		}
		parts := strings.Split(strings.TrimSpace(m.Text), "\n\n")
		if m.Edit == 0 && m.Del == 0 && (m.Ask || strings.Contains(parts[len(parts)-1], "?")) {
			out = append(out, id)
		}
	}
	checked(rows.Err())
	return out
}
