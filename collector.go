package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"./hone"
)

func main() {
	// command-line flags
	debugFlag := flag.Bool("daemon", false, "Daemonize")
	flag.Parse()

	// create collector
	collector := new(hone.Collector)

	cleanup := func() {
		fmt.Printf("cleaning up\n")
		collector.Stop()
	}
	
	// signal handler
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT)
	go func () {
		sig := <-sigChan
		fmt.Printf("got signal %s\n", sig)
		cleanup()
		os.Exit(0)
	}()


	// daemonize?
	fmt.Printf("Starting collector... %#v\n", collector)
	if *debugFlag {
		fmt.Println("Daemonizing...")
	} else {
		fmt.Println("Runnning in foreground...")
	}

	// event handler
	go runCollector(collector)
	
	// block forever
	select {}
}

func runCollector(collector *hone.Collector) {
	// run collector
	eventChan := collector.Run()

	for {
		evt := <-eventChan

		
		fmt.Printf("Got evt: %#v\n", evt)
	}
}
