package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMediaAgentsAndTopics(t *testing.T) {
	s := openServer(t.TempDir())
	defer s.db.Close()
	request := func(key, a string, body io.Reader, ct string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", "/api.php?"+a, body)
		r.Header.Set("X-Key", key)
		r.Header.Set("Content-Type", ct)
		w := httptest.NewRecorder()
		s.api(w, r)
		if w.Code != status {
			t.Fatalf("%s returned %d: %s", a, w.Code, w.Body.String())
		}
		return w
	}
	form := func(key, a string, v url.Values, status int) obj {
		t.Helper()
		w := request(key, a, strings.NewReader(v.Encode()), "application/x-www-form-urlencoded", status)
		var out obj
		if e := json.Unmarshal(w.Body.Bytes(), &out); e != nil {
			t.Fatal(e)
		}
		return out
	}
	owner := form("", "a=setup", url.Values{"name": {"Аня"}}, 200)["key"].(string)
	guest := form(owner, "a=invite", url.Values{"name": {"Боря"}}, 200)["key"].(string)
	form(owner, "a=new", url.Values{"title": {"тема"}, "members": {"Боря"}}, 200)
	key := func(sub string) string { return form(owner, "a=aikey", url.Values{"sub": {sub}}, 200)["key"].(string) }
	agent := key("код")
	other := key("тест")
	if key("код") != agent {
		t.Fatal("unstable active key")
	}
	var body bytes.Buffer
	mp := multipart.NewWriter(&body)
	f, e := mp.CreateFormFile("file[]", "photo.png")
	if e != nil {
		t.Fatal(e)
	}
	var encoded bytes.Buffer
	if e := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, 2, 2))); e != nil {
		t.Fatal(e)
	}
	png := encoded.Bytes()
	f.Write(png)
	mp.WriteField("priv", "1")
	mp.Close()
	request(owner, "a=post&t=1", &body, mp.FormDataContentType(), 200)
	get := func() []any { return form(owner, "a=read&t=1", nil, 200)["messages"].([]any) }
	filename := get()[0].(map[string]any)["files"].([]any)[0].(map[string]any)["f"].(string)
	path := "a=file&t=1&f=" + filename
	request(guest, path, nil, "", 404)
	if !bytes.Equal(request(owner, path, nil, "", 200).Body.Bytes(), png) {
		t.Fatal("file mismatch")
	}
	cookie := request(owner, "a=me", nil, "", 200).Result().Cookies()[0]
	r := httptest.NewRequest(http.MethodGet, "/api.php?"+path, nil)
	r.AddCookie(cookie)
	w := httptest.NewRecorder()
	s.api(w, r)
	if w.Code != 200 {
		t.Fatal("media cookie")
	}
	form(agent, "a=post&t=1", url.Values{"text": {"ответ"}}, 200)
	form(other, "a=edit&t=1", url.Values{"on": {"2"}, "text": {"чужое"}}, 403)
	form(agent, "a=edit&t=1", url.Values{"on": {"2"}, "text": {"правка"}}, 200)
	form(agent, "a=del&t=1", url.Values{"on": {"2"}}, 200)
	form(owner, "a=del&t=1", url.Values{"on": {"1"}}, 200)
	request(owner, path, nil, "", 404)
	form(owner, "a=rename&t=1", url.Values{"title": {"новая"}}, 200)
	form(owner, "a=archive&t=1", nil, 200)
	topic := form(owner, "a=read&t=1", nil, 200)["topic"].(map[string]any)
	if topic["title"] != "новая" || topic["archived"] != true {
		t.Fatal(topic)
	}
	form(owner, "a=aikey", url.Values{"sub": {"код"}, "off": {"1"}}, 200)
	form(agent, "a=me", nil, 401)
	if key("код") == agent {
		t.Fatal("revoked key reused")
	}
	form(guest, "a=typing&t=1", nil, 200)
	form(other, "a=usage&t=1", url.Values{"u": {"since start\nclaude-sonnet-4 1000000 0 0 0 0"}}, 200)
	v := form(owner, "a=read&t=1", nil, 200)
	if len(v["typing"].([]any)) != 1 || v["usage"].([]any)[0].(map[string]any)["usd"] != float64(3) {
		t.Fatal("live data")
	}
	if holiday(time.Date(2026, 11, 29, 0, 0, 0, 0, time.UTC))["name"] != "День матери" {
		t.Fatal("calendar")
	}
}
