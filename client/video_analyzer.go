package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"strings"
)

type VideoAnalyzer struct {
	sshHost         string
	sshUser         string
	inputDirectory  string
	outputDirectory string
	inputCh         chan string
	cmdVmz          *exec.Cmd
	cmdVmzIn        io.WriteCloser
	cmdVmzOut       io.ReadCloser
	cmdVmzOutReader *bufio.Reader
	cmdVmzErr       io.ReadCloser
}

func NewVideoAnalyzer(sshUser string, sshHost string, inputDirectory string, outputDirectory string, inputCh chan string) *VideoAnalyzer {
	return &VideoAnalyzer{
		sshUser:         sshUser,
		sshHost:         sshHost,
		inputDirectory:  inputDirectory,
		outputDirectory: outputDirectory,
		inputCh:         inputCh,
	}
}

func (a *VideoAnalyzer) PrepareVMZ() error {
	a.cmdVmz = exec.Command("ssh", fmt.Sprintf("%s@%s", a.sshUser, a.sshHost), ". /etc/profile && cd VMZ && . demo-0-initialize.sh && ./demo-1-analyze.sh ../tx2test.mp4")
	var err error
	a.cmdVmzIn, err = a.cmdVmz.StdinPipe()
	if err != nil {
		return err
	}
	a.cmdVmzOut, err = a.cmdVmz.StdoutPipe()
	if err != nil {
		return err
	}
	a.cmdVmzErr, err = a.cmdVmz.StderrPipe()
	if err != nil {
		return err
	}
	err = a.cmdVmz.Start()
	if err != nil {
		return err
	}
	a.cmdVmzOutReader = bufio.NewReader(a.cmdVmzOut)
	err = a.waitVmzReady()
	if err != nil {
		return err
	}
	return nil
}

func (a *VideoAnalyzer) waitVmzReady() error {
	for {
		line, _, err := a.cmdVmzOutReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				stderrStr, _ := ioutil.ReadAll(a.cmdVmzErr)
				return fmt.Errorf("EOF: stderr is: %s", stderrStr)
			}
			return err
		}
		lineStr := string(line)
		// log.Println(lineStr)
		if lineStr == "EXTRACT-FEATURES-READY-MARKER" {
			return nil
		}
	}
}

func (a *VideoAnalyzer) Next() error {
	filename := <-a.inputCh
	log.Println("Sending", filename)
	scpMovieCmd := exec.Command("scp", path.Join(a.inputDirectory, filename), fmt.Sprintf("%s@%s:tx2test.mp4", a.sshUser, a.sshHost))
	err := scpMovieCmd.Run()
	if err != nil {
		return err
	}

	log.Println("Running VMZ on", filename)
	_, err = a.cmdVmzIn.Write([]byte("\n"))
	if err != nil {
		return err
	}
	err = a.waitVmzReady()
	if err != nil {
		return err
	}

	log.Println("Generating SRT for", filename)
	outputCmd := exec.Command("ssh", fmt.Sprintf("%s@%s", a.sshUser, a.sshHost), ". /etc/profile && cd VMZ && . demo-0-initialize.sh && ./demo-2-output.py ../tx2test.mp4")
	err = outputCmd.Run()
	if err != nil {
		return err
	}

	log.Println("Receiving SRT", filename)
	subtitlePath := path.Join(a.outputDirectory, strings.TrimSuffix(filename, path.Ext(filename))+".srt")
	scpSubtitleCmd := exec.Command("scp", fmt.Sprintf("%s@%s:my_features_softmax.srt", a.sshUser, a.sshHost), subtitlePath)
	err = scpSubtitleCmd.Run()
	if err != nil {
		return err
	}

	log.Println("Burning substile", filename)
	ffmpegSubtitleCmd := exec.Command("ffmpeg", "-i", path.Join(a.inputDirectory, filename), "-vf", fmt.Sprintf("subtitles=%s", subtitlePath), path.Join(a.outputDirectory, filename))
	err = ffmpegSubtitleCmd.Run()
	if err != nil {
		return err
	}

	return nil
}
