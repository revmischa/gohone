package main

import (
	"./hone"
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
	go agent.Run()
	logger.Info("Agent started")
}
