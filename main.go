package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"time"
)

//go:embed web/*
var assets embed.FS

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
	log.Fatal((&http.Server{Addr: *addr, Handler: s.routes(), ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}
func (s *server) routes() http.Handler {
	static, err := fs.Sub(assets, "web")
	checked(err)
	mux := http.NewServeMux()
	mux.HandleFunc("/api.php", s.api)
	mux.Handle("/", http.FileServer(http.FS(static)))
	return mux
}
