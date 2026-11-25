package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestImportPHP(t *testing.T) {
	src := t.TempDir()
	write := func(n, data string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(src, n+".json.php"), []byte("<?php exit; ?>\n"+data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("keys", fmt.Sprintf("{%q:%q}", digest("test"), "Аня"))
	write("admins", `["Аня"]`)
	write("topics", `[{"id":1,"by":"Аня"},{"id":2,"by":"Аня","members":["Аня"]}]`)
	write("t1", `[{"id":1,"who":"Аня","text":"раз"}]`)
	s := openServer(t.TempDir())
	defer s.db.Close()
	if s.importPHP(src) == nil {
		t.Fatal("accepted missing history")
	}
	write("t2", `[{"id":1,"who":"Аня","text":"два"}]`)
	if err := s.importPHP(src); err != nil {
		t.Fatal(err)
	}
	if id := s.appendMessage(2, obj{"text": "три"}); id != 2 {
		t.Fatal(id)
	}
	if s.importPHP(src) == nil {
		t.Fatal("accepted repeated import")
	}
}
