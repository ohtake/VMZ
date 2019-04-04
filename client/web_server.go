package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
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

type contentHandler struct {
	s   *WebServer
	dir string
	ext string
}

func (h contentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	contentFileName := strings.TrimSuffix(h.s.filenames[id], path.Ext(h.s.filenames[id])) + "." + h.ext
	contentFilePath := path.Join(h.dir, contentFileName)
	contentFile, err := os.Open(contentFilePath)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
	}
	defer contentFile.Close()
	stat, err := contentFile.Stat()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err)
	}
	http.ServeContent(w, r, contentFile.Name(), stat.ModTime(), contentFile)
}

func (s *WebServer) Serve() {
	http.Handle("/video", contentHandler{s: s, dir: s.inputDirectory, ext: "mp4"})
	http.Handle("/actions", contentHandler{s: s, dir: s.outputDirectory, ext: "csv"})
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}
