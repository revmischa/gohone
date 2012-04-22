package hone

import (
	"bufio"
	"log"
	"os"
	"fmt"
	"regexp"
	"strconv"
)

const (
	capFilePath = "/dev/honet"
)

// receiver of capture events
type EventChan chan *CaptureEvent

type CaptureEventType string

// a capture event
type CaptureEvent struct {
	Type CaptureEventType
	Packet []byte

	HostGUID string

	ConncetionState rune
	Direction rune
	Sockfd uint64
	Proto string
	Src string
	Dst string
	Len uint64

	PID uint64
	PPID uint64
	TGID uint64
	UID uint64
	GID uint64

	Executable string
	Argv string
}

// capture event type

// class that reads and dispatches capture events
type Collector struct {
	CaptureFile *os.File
	Stopped bool
	EventCount uint64
}

// open kernel hone event module
func (collector *Collector) OpenCaptureFile() {
	var err error
	collector.CaptureFile, err = os.Open(capFilePath)

	// TODO: check if module is loaded

	if err != nil {
		log.Fatalf("Error opening capture file for reading: %s\n", err)
	}
}

func (collector *Collector) CloseCaptureFile() {
	collector.CaptureFile.Close()
}

// opens capture, starts sending events to event channel
func (collector *Collector) Run() EventChan {
	// open capture
	collector.OpenCaptureFile()

	// create a channel to receive capture events
	ec := make(EventChan)

	// start reading file
	go func() {
		for (! collector.Stopped) {
			reader := bufio.NewReader(collector.CaptureFile)
			line, isPrefix, err := reader.ReadLine()
			if err != nil {
				log.Panicf("Error reading capture file: %s\n", err)
				continue
			}

			// partial line? means line is 4k long. not chill.
			if isPrefix {
				log.Panicf("Error finding end-of-line in capture")
				continue
			}

			collector.EventCount++

			// parse input
			evt := collector.ParseHoneEventLine(line)
			if (evt == nil) {
				continue
			}

			// success
			ec <- evt
		}

		collector.CloseCaptureFile()
	}()

	return ec
}

func (collector *Collector) Stop() {
	collector.Stopped = true
}

// parses a line from /dev/honet into a CaptureEvent
func (collector *Collector) ParseHoneEventLine(lineBytes []byte) *CaptureEvent {
	line := string(lineBytes)

	parseSuccess := false

	//log.Printf("line: %s\n", line)
	
	// parse timestamp and event type
	evt := new(CaptureEvent)
	var delta float64
	var eventType CaptureEventType
	parsed, err := fmt.Sscanf(line, "%f %s", &delta, &eventType)
	if (err == nil && parsed == 2) {
		// handle event types
		//procSpec := "%d %d %d %d %d"
		procSpec := "(\\d+) (\\d+) (\\d+) (\\d+) (\\d+)"
		switch (eventType) {
		case "EXEC", "EXIT", "FORK":
			// process event
			re := regexp.MustCompile(procSpec + "(?: \"([^\"]+)\" (.+))?")
			matches := re.FindStringSubmatch(line)

			if (len(matches) >= 6) {
				evt.PID  = parseUint(matches[1])
				evt.PPID = parseUint(matches[2])
				evt.TGID = parseUint(matches[3])
				evt.UID  = parseUint(matches[4])
				evt.GID  = parseUint(matches[5])
				parseSuccess = true
			}
			
			if (parseSuccess && eventType == "EXEC") {
				evt.Executable = matches[6]
				evt.Argv = matches[7]
			}			
			
		case "PAKT":
			// packet
			re := regexp.MustCompile("([IO]) ([A-Fa-f0-9]+) (\\S+) (\\S+) -> (\\S+) (\\d+)")
			matches := re.FindStringSubmatch(line)

			if (len(matches) == 7) {
				evt.Direction = rune(matches[1][0])
				evt.Sockfd, _ = strconv.ParseUint(matches[2], 16, 0)
				evt.Proto = matches[3]
				evt.Src = matches[4]
				evt.Dst = matches[5]
				evt.Len = parseUint(matches[6])
				parseSuccess = true
			}

		case "SOCK":
			// socket
			re := regexp.MustCompile("([CO]) " + procSpec + " ([A-Fa-f0-9]+)")
			matches := re.FindStringSubmatch(line)

			if (len(matches) == 8) {
				evt.Direction = rune(matches[1][0])
				evt.Sockfd, _ = strconv.ParseUint(matches[7], 16, 0)

				evt.PID  = parseUint(matches[2])
				evt.PPID = parseUint(matches[3])
				evt.TGID = parseUint(matches[4])
				evt.UID  = parseUint(matches[5])
				evt.GID  = parseUint(matches[6])
				parseSuccess = true
			}

		case "HEAD":
			// capture header, host GUID
			_, err := fmt.Sscanf(line, "%f %s %s", &delta, &eventType, &evt.HostGUID)
			if (err != nil) {
				log.Printf("Failed to parse HEAD event: %s\n", err)
			} else {
				parseSuccess = true
			}
			
		default:
			if (collector.EventCount > 10) {
				log.Printf("unhandled hone event type: %s\n", eventType)
			}
		}
	}

	if (! parseSuccess) {
		if (collector.EventCount > 10) {
			log.Printf("Failed to parse line '%s': %s\n", line, err)
		}
	}

	if (parseSuccess) {
		// event is chill
		//evt.Packet = lineBytes
		evt.Type = eventType
		return evt
	}

	return nil
}

func parseUint(s string) uint64 {
	res, err := strconv.ParseUint(s, 10, 0)
	if (err != nil) {
		log.Printf("failed to convert '%s' to int\n", s)
	}
	
	return res;
}