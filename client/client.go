package main

import (
	"flag"
	"fmt"
	"log"
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
	if *help {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("%s (%d sec) -> %s\n", *inputVideoPath, *watchInterval, *outputVideoPath)

	inputCh := make(chan string, 10)
	analyzedCh := make(chan string, 10)

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

	var videoAnalyzer *VideoAnalyzer
	if *sshUser == "" || *sshHost == "" {
		log.Println("SSH is not configured. Disabling analyzer.")
		videoAnalyzer = NewVideoAnalyzerFake(*inputVideoPath, *outputVideoPath, inputCh, analyzedCh)
	} else {
		videoAnalyzer = NewVideoAnalyzer(*sshUser, *sshHost, *inputVideoPath, *outputVideoPath, inputCh, analyzedCh)
		err := videoAnalyzer.PrepareVMZ()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failure to prepare VMZ")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}

	go func() {
		for {
			err := videoAnalyzer.Next()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				// TODO retry?
			}
		}
	}()

	webServer := NewWebServer(8080, *inputVideoPath, *outputVideoPath, analyzedCh)
	go func() {
		for {
			webServer.AddAnalyzed()
		}
	}()
	webServer.Serve()
}
