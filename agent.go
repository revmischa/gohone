package main

import (
	"./hone"
	"encoding/json"
	"flag"
	"fmt"
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
		fmt.Printf("Error connecting to syslog: %s\n", err)
		os.Exit(1)
	}

	// command-line flags
	serverAddr := flag.String("server", "", "Destination server address")
	serverPort := flag.Uint("port", 7100, "Destination server port")
	flag.Parse()

	if len(*serverAddr) == 0 {
		fmt.Println("Destination server is required")
		os.Exit(1)
	}

	// create agent
	agent := hone.NewAgent(*serverAddr, *serverPort)

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

	go agent.Connect()

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
		handleEvent(agent, evt)
	}
}

func handleEvent(agent *hone.Agent, evt *hone.CaptureEvent) {
	if !agent.ConnectedToServer {
		go agent.Connect()
		return
	}

	// we don't want packets generated by this process (lol!)
	if evt.TGID != 0 && agent.CommunicationPID != 0 {
		ourPID := agent.CommunicationPID

		if evt.TGID == ourPID {
			// discard
			return
		}
	}

	// encode as JSON
	jsonStr, err := json.Marshal(evt)
	if err != nil {
		logger.Err(err.Error())
		return
	}

	//fmt.Println(string(jsonStr))
	//fmt.Printf("Got evt: %#v\n", evt)

	// send newline-terminated JSON event to server
	fmt.Fprintln(agent.ServerConn, string(jsonStr))
}
