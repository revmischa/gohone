package main

import (
	"flag"
	"fmt"
	"hone"
	"hone/reader/honet"
	"hone/reader/ntar"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"
)

var logger *syslog.Writer

func main() {
	var err error
	logger, err = syslog.New(syslog.LOG_DEBUG, "hone-agent")
	if err != nil {
		log.Panicf("Error connecting to syslog: %s\n", err)
	}

	// command-line flags
	serverAddr := flag.String("server", "", "Destination server address")
	serverPort := flag.Uint("port", 7100, "Destination server port")
	sourceModule := flag.String("source", "ntar", "Hone event source (ntar, honet)")
	flag.Parse()

	if len(*serverAddr) == 0 {
		log.Fatalln("Destination server is required")
	}

	// create agent
	var reader hone.Reader
	switch *sourceModule {
	case "ntar":
		reader = ntar.NewReader()
	case "honet":
		reader = honet.NewReader()
	default:
		log.Fatalf("Unknown source module %s\n")
	}

	agent := hone.NewAgent(*serverAddr, *serverPort, reader)

	cleanup := func() {
		logger.Debug("Cleaning up and exiting")
		agent.Stop()
	}

	// signal handler
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		fmt.Printf("got signal %s\n", sig)
		cleanup()
		os.Exit(0)
	}()

	eventChan := agent.Start()
	logger.Info("Agent started")
	go agent.Run()

	// block forever
	for {
		<-eventChan
		//fmt.Printf("got event %s\n", evt)
	}
}
