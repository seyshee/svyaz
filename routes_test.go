package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedRoutes(t *testing.T) {
	s := openServer(t.TempDir())
	defer s.db.Close()
	handler := s.routes()
	for _, tc := range []struct {
		path, content string
		status        int
	}{
		{"/", `src="app.js"`, 200}, {"/", `href="styles.css"`, 200}, {"/app.js", "function api(", 200}, {"/styles.css", ":root", 200}, {"/kak.md", "a=wait", 200}, {"/api.php?a=me", `"setup":true`, 401}, {"/store.go", "", 404},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.content) {
			t.Errorf("unexpected response for %s (status %d)", tc.path, w.Code)
		}
	}
}
