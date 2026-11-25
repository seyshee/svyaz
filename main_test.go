package main

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPrivateConversation(t *testing.T) {
	s := openServer(t.TempDir())
	defer s.db.Close()
	call := func(key, query, body string, status int) obj {
		req := httptest.NewRequest("POST", "/api.php?"+query, strings.NewReader(body))
		req.Header.Set("X-Key", key)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		s.api(w, req)
		if w.Code != status {
			t.Fatalf("%s: HTTP %d", query, w.Code)
		}
		var v obj
		if status != 204 {
			if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
				t.Fatal(err)
			}
		}
		return v
	}
	owner := call("", "a=setup", "name=Анна", 200)["key"].(string)
	guest := call(owner, "a=invite", "name=Боря", 200)["key"].(string)
	other := call(owner, "a=invite", "name=Иван", 200)["key"].(string)
	ai := call(guest, "a=aikey", "", 200)["key"].(string)
	call(owner, "a=new", "title=чат&members=Боря", 200)
	call(other, "a=read&t=1", "", 404)
	call(owner, "a=post&t=1", url.Values{"text": {"вопрос"}, "to": {"ИИ Бори"}}.Encode(), 200)
	if len(call(ai, "a=wait&t=1&wait=0", "", 200)["messages"].([]any)) != 1 {
		t.Fatal("missing question")
	}
	call(ai, "a=post&t=1", "text=ответ&reply=1", 200)
	msgs := call(owner, "a=read&t=1", "", 200)["messages"].([]any)
	if len(msgs) != 2 || msgs[1].(map[string]any)["from"] != "ИИ Бори" {
		t.Fatal("missing reply")
	}
	call(ai, "a=wait&t=1&since=1&wait=0", "", 204)
	call(owner, "a=post&t=1", "text=секрет&priv=1", 200)
	if len(call(guest, "a=read&t=1&since=2", "", 200)["messages"].([]any)) != 0 {
		t.Fatal("private message disclosed")
	}
	call(guest, "a=react&t=1", "on=3&r=agree", 404)
	call(owner, "a=react&t=1", "on=2&r=agree", 200)
	got := call(ai, "a=wait&t=1&since=3&wait=0", "", 200)["messages"].([]any)
	if len(got) != 1 || got[0].(map[string]any)["meaning"] != "согласен" {
		t.Fatal("missing reaction")
	}
	call(owner, "a=react&t=1", "on=2&r=agree&off=1", 200)
	call(owner, "a=task&t=1", "title=x&ai=1", 403)
	call(other, "a=tasks&t=1", "", 404)
	call(ai, "a=task&t=1", "title=x&prio=2", 200)
	call(ai, "a=task&t=1", "n=1&state=done&msg=2", 200)
	req := httptest.NewRequest("GET", "/api.php?a=tasks&t=1", nil)
	req.Header.Set("X-Key", owner)
	w := httptest.NewRecorder()
	s.api(w, req)
	var tasks []struct {
		State string
		Log   []obj
	}
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil || len(tasks) != 1 || tasks[0].State != "done" || len(tasks[0].Log) != 2 || tasks[0].Log[1]["msg"] != float64(2) {
		t.Fatalf("unexpected task history: %v", tasks)
	}
	call(owner, "a=post&t=1", "text=Резюме:+итог", 200)
	call(ai, "a=post&t=1", "text=ответ", 200)
	if len(call(ai, "a=read&t=1&from=resume", "", 200)["messages"].([]any)) != 2 {
		t.Fatal("resume")
	}
	page := call(owner, "a=read&t=1&tail=1", "", 200)
	if len(page["messages"].([]any)) != 1 || page["skip"].(float64) < 1 {
		t.Fatal("pagination")
	}
	call(owner, "a=read&t=1&tail=200&before=2", "", 200)
	call(ai, "a=add&t=1", "who=Иван", 403)
	call(owner, "a=add&t=1", "who=Иван", 200)
	call(other, "a=kick&t=1", "who=Боря", 403)
	result := make(chan int, 1)
	go func() {
		req := httptest.NewRequest("GET", "/api.php?a=wait&t=1&since=9999&wait=10", nil)
		req.Header.Set("X-Key", ai)
		w := httptest.NewRecorder()
		s.api(w, req)
		result <- w.Code
	}()
	call(owner, "a=kick&t=1", "who=Боря&why=test", 200)
	if <-result != 410 {
		t.Fatal("pending wait allowed after kick")
	}
	call(owner, "a=add&t=1", "who=Боря", 200)
	call(guest, "a=read&t=1", "", 200)
	call(owner, "a=forget", "who=Боря", 200)
	call(guest, "a=me", "", 401)
	call(ai, "a=read&t=1", "", 401)
}
