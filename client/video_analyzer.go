package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
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
	for {
		var line []byte
		line, _, err = a.cmdVmzOutReader.ReadLine()
		if err != nil {
			if err == io.EOF {
				stderrStr, _ := ioutil.ReadAll(a.cmdVmzErr)
				return fmt.Errorf("EOF: stderr is: %s", stderrStr)
			}
			return err
		}
		lineStr := string(line)
		if lineStr == "Press enter if you place a video file:" {
			return nil
		}
		// log.Println(lineStr)
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
