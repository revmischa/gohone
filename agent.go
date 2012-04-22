package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"os/signal"
	"syscall"
	"encoding/json"
	"./hone"
)

var logger *syslog.Writer

func main() {
	var err error
	logger, err = syslog.New(syslog.LOG_DEBUG, "hone-agent")
	if (err != nil) {
		fmt.Printf("Error connecting to syslog: %s\n", err)
		os.Exit(1)
	}
	
	// command-line flags
	serverAddr := flag.String("server", "", "Destination server address")
	flag.Parse()

	if (serverAddr == nil) {
		fmt.Println("Destination server is required")
		os.Exit(1)
	}

	// create agent
	agent := new(hone.Agent)

	cleanup := func() {
		logger.Debug("Cleaning up and exiting")
		agent.Stop()
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

	// event handler
	go runAgent(agent)
	
	// block forever
	select {}
}

func runAgent(agent *hone.Agent) {
	// run agent
	eventChan := agent.Run()

	logger.Info("Agent started")

	for {
		evt := <-eventChan
		handleEvent(evt)
	}
}

func handleEvent(evt *hone.CaptureEvent) {
	//fmt.Printf("Got evt: %#v\n", evt)

	// encode as JSON
	jsonStr, err := json.Marshal(evt)
	if (err != nil) {
		logger.Err(err.Error())
		return
	}
	
	fmt.Println(string(jsonStr))
}
