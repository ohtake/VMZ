package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
)

const staticRoot = "web/"

type WebServer struct {
	port            int
	inputDirectory  string
	outputDirectory string
	filenames       []string
	analyzedCh      chan string
}

func NewWebServer(port int, inputDirectory string, outputDirectory string, analyzedCh chan string) *WebServer {
	return &WebServer{
		port:            port,
		inputDirectory:  inputDirectory,
		outputDirectory: outputDirectory,
		analyzedCh:      analyzedCh,
	}
}

func (s *WebServer) AddAnalyzed() {
	filename := <-s.analyzedCh
	s.filenames = append(s.filenames, filename)
}

type videoHandler struct {
	s *WebServer
}

func (h videoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var id int
	if _, err := fmt.Sscan(r.URL.Query().Get("id"), &id); err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err)
		return
	}
	if id < 0 {
		w.WriteHeader(400)
		fmt.Fprint(w, "id must not be negative")
		return
	}
	if id >= len(h.s.filenames) {
		w.WriteHeader(404)
		fmt.Fprint(w, "not analyzed: ", id)
		return
	}

	tsFile, err := os.Open(path.Join(h.s.inputDirectory, h.s.filenames[id]))
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
	}
	defer tsFile.Close()
	stat, err := tsFile.Stat()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
	}
	http.ServeContent(w, r, tsFile.Name(), stat.ModTime(), tsFile)
}

func (s *WebServer) Serve() {
	http.Handle("/video", videoHandler{s: s})
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}
