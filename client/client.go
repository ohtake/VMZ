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
	help := flag.Bool("help", false, "Print help")
	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("%s (%d sec) -> %s\n", *inputVideoPath, *watchInterval, *outputVideoPath)

	inputCh := make(chan string, 10)
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
		videoAnalyzer := NewVideoAnalyzer(*inputVideoPath, *outputVideoPath, inputCh)
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
