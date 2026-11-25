package main

import "database/sql"

const mid = "coalesce(json_extract(body,'$.id'),id)"
const visible = " FROM messages WHERE t=? AND (coalesce(json_extract(body,'$.priv'),0)=0 OR json_extract(body,'$.who')=?)"
const ordinary = " AND json_extract(body,'$.react') IS NULL AND json_extract(body,'$.edit') IS NULL AND json_extract(body,'$.del') IS NULL"

func (s *server) history(t int, me string, arg func(string) string) (*sql.Rows, int, int, error) {
	since := number(arg("since"))
	before := number(arg("before"))
	skip, j := 0, 0
	if tail := number(arg("tail")); arg("a") == "read" && tail > 0 {
		var first sql.NullInt64
		e := s.db.QueryRow("SELECT "+mid+visible+ordinary+" AND (?=0 OR "+mid+"<?) ORDER BY "+mid+" DESC LIMIT 1 OFFSET ?", t, me, before, before, tail-1).Scan(&first)
		if e != sql.ErrNoRows {
			checked(e)
		}
		since = 0
		if first.Valid {
			since = int(first.Int64) - 1
		}
		checked(s.db.QueryRow("SELECT count(*)"+visible+" AND "+mid+"<=?", t, me, since).Scan(&skip))
		checked(s.db.QueryRow("SELECT count(*)"+visible+ordinary+" AND NOT coalesce(json_extract(body,'$.sys'),0) AND NOT coalesce(json_extract(body,'$.gone'),0) AND "+mid+"<=? AND "+mid+">coalesce((SELECT max("+mid+")"+visible+" AND "+mid+"<=? AND json_extract(body,'$.sys') AND (json_extract(body,'$.text') LIKE '%позвал%' OR json_extract(body,'$.text') LIKE '%убрал(а)%')),0)", t, me, since, t, me, since).Scan(&j))
	} else if arg("from") == "resume" {
		var last sql.NullInt64
		checked(s.db.QueryRow("SELECT max("+mid+")"+visible+ordinary+" AND ltrim(json_extract(body,'$.text'),' *') LIKE 'Резюме:%'", t, me).Scan(&last))
		if last.Valid {
			since = int(last.Int64) - 1
		}
	}
	rows, e := s.db.Query("SELECT "+mid+",body"+visible+" AND "+mid+">? AND (?=0 OR "+mid+"<?) ORDER BY "+mid, t, me, since, before, before)
	return rows, skip, j, e
}
