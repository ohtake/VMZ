package main

import (
	"io/ioutil"
	"path"
	"sort"
	"strings"
)

type InputWatcher struct {
	directory    string
	lastFilename string
	inputCh      chan string
}

func NewInputWatcher(directory string, inputCh chan string) *InputWatcher {
	return &InputWatcher{
		directory: directory,
		inputCh:   inputCh,
	}
}

func (w *InputWatcher) Check() error {
	fis, err := ioutil.ReadDir(w.directory)
	if err != nil {
		return err
	}
	var newFiles []string
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		if path.Ext(fi.Name()) != ".mp4" {
			continue
		}
		if w.lastFilename == "" || strings.Compare(w.lastFilename, fi.Name()) < 0 {
			newFiles = append(newFiles, fi.Name())
		}
	}
	if len(newFiles) <= 1 {
		return nil
	}
	sort.Strings(newFiles)
	// Skip latest file because it may not be finised
	newFiles = newFiles[:len(newFiles)-1]
	for _, filename := range newFiles {
		w.inputCh <- filename
	}
	w.lastFilename = newFiles[len(newFiles)-1]
	return nil
}
