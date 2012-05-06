package main

/*
import (
	"hone"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	// create collector
	collector := new(hone.Collector)

	cleanup := func() {
		fmt.Printf("cleaning up\n")
		collector.Stop()
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

	// block forever
	select {}
}
*/
