package main

import "fmt"

type VideoAnalyzer struct {
	inputDirectory  string
	outputDirectory string
	inputCh         chan string
}

func NewVideoAnalyzer(inputDirectory string, outputDirectory string, inputCh chan string) *VideoAnalyzer {
	return &VideoAnalyzer{
		inputDirectory:  inputDirectory,
		outputDirectory: outputDirectory,
		inputCh:         inputCh,
	}
}

func (a *VideoAnalyzer) Next() error {
	filename := <-a.inputCh
	fmt.Println(filename)
	// TODO prepare VMZ (how?)
	// TODO scp mp4 to GPU server
	// TODO run VMZ (how?)
	// TODO scp srt from GPU server
	return nil
}
