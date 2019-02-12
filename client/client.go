package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	inputVideoPath := flag.String("input", "input", "Path of the input directory")
	outputVideoPath := flag.String("output", "output", "Path of the output directory")
	watchInterval := flag.Int("interval", 10, "Interval of watching input directory (sec)")
	sshUser := flag.String("ssh-user", "", "user name for ssh connection to VMZ host")
	sshHost := flag.String("ssh-host", "", "host name for ssh connection to VMZ host")
	help := flag.Bool("help", false, "Print help")
	flag.Parse()
	if *help || *sshUser == "" || *sshHost == "" {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("%s (%d sec) -> %s\n", *inputVideoPath, *watchInterval, *outputVideoPath)

	inputCh := make(chan string, 10)
	videoAnalyzer := NewVideoAnalyzer(*sshUser, *sshHost, *inputVideoPath, *outputVideoPath, inputCh)
	err := videoAnalyzer.PrepareVMZ()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failure to prepare VMZ")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	go func() {
		inputWatcher := NewInputWatcher(*inputVideoPath, inputCh)
		for {
			err := inputWatcher.Check()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			time.Sleep(time.Duration(*watchInterval) * time.Second)
		}
	}()

	go func() {
		for {
			err := videoAnalyzer.Next()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				// TODO retry?
			}
		}
	}()
	select {}
}
