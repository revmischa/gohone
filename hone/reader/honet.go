package honet

import (
	"../../hone"
	"regexp"
	"os"
	"log"
	"fmt"
	"strconv"
)


func NewReader() hone.Reader {
	return hone.Reader(new(Reader))
}

type Reader struct {}

// open kernel hone event module
func (*Reader) OpenCaptureFile(agent *hone.Agent, capFilePath string) {
	var err error
	agent.CaptureFile, err = os.Open(capFilePath)

	// TODO: check if module is loaded
	
	if err != nil {
		log.Fatalf("Error opening capture file for reading: %s\n", err)
	}
}

func (*Reader) CloseCaptureFile(agent *hone.Agent) {
	agent.CaptureFile.Close()
}

// parses a line from /dev/honet into a CaptureEvent
func (*Reader) ParseHoneEventLine(agent *hone.Agent, lineBytes []byte) *hone.CaptureEvent {
	line := string(lineBytes)

	parseSuccess := false

	//log.Printf("line: %s\n", line)

	// parse timestamp and event type
	var delta float64
	var eventType hone.CaptureEventType
	parsed, err := fmt.Sscanf(line, "%f %s", &delta, &eventType)

	if err != nil || parsed != 2 {
		if agent.EventCount > 10 {
			// first few lines might be incomplete
			log.Printf("Failed to parse line '%s': %s\n", line, err)
		}
		return nil
	}
	
	// build event
	evt := new(hone.CaptureEvent)
	evt.CaptureTimeDelta = delta
	
	// handle event types
	procSpec := "(\\d+) (\\d+) (\\d+) (\\d+) (\\d+)"
	switch eventType {
	case "EXEC", "EXIT", "FORK":
		// process event
		re := regexp.MustCompile(procSpec + "(?: \"([^\"]+)\" (.+))?")
		matches := re.FindStringSubmatch(line)

		if len(matches) >= 6 {
			evt.PID = parseInt(matches[1])
			evt.PPID = parseInt(matches[2])
			evt.TGID = parseInt(matches[3])
			evt.UID = parseInt(matches[4])
			evt.GID = parseInt(matches[5])
			parseSuccess = true
		}

		if parseSuccess && eventType == "EXEC" {
			evt.Executable = matches[6]
			evt.Args = matches[7]

			agent.ExecEvents[evt.TGID] = evt
		}
		
		if parseSuccess && eventType == "EXIT" {
			delete(agent.ExecEvents, evt.TGID)
		}

	case "PAKT":
		// packet
		re := regexp.MustCompile("([IO]) ([A-Fa-f0-9]+) (\\S+) (\\S+) -> (\\S+) (\\d+)")
		matches := re.FindStringSubmatch(line)

		if len(matches) != 7 {
			fmt.Printf("failed to parse PAKT evt '%s'\n", line)
			return nil
		}
		
		evt.Direction = matches[1]
		evt.Sockfd, _ = strconv.ParseUint(matches[2], 16, 0)
		evt.Proto = matches[3]
		evt.Src = matches[4]
		evt.Dst = matches[5]

		evtlen, err := strconv.ParseUint(matches[6], 10, 0)
		if err == nil {
			evt.Len = evtlen
			parseSuccess = true
		} else {
			fmt.Printf("Failed to parse length %s: %s\n", matches[6], err)
		}
		
	case "SOCK":
		// socket
		re := regexp.MustCompile("([CO]) " + procSpec + " ([A-Fa-f0-9]+)")
		matches := re.FindStringSubmatch(line)

		if len(matches) == 8 {
			evt.ConnectionState = matches[1]
			evt.Sockfd, _ = strconv.ParseUint(matches[7], 16, 0)

			evt.PID = parseInt(matches[2])
			evt.PPID = parseInt(matches[3])
			evt.TGID = parseInt(matches[4])
			evt.UID = parseInt(matches[5])
			evt.GID = parseInt(matches[6])

			if evt.PID != 0 && evt.Sockfd != 0 {
				parseSuccess = true
				agent.SockEvents[evt.Sockfd] = evt
			} else {
				fmt.Printf("Failed to find PID/sockfd from SOCK\n");
			}
		} else {
			fmt.Printf("Failed to parse SOCK event: '%s'\n", line)
		}

	case "HEAD":
		// capture header, host GUID
		_, err := fmt.Sscanf(line, "%f %s %s", &delta, &eventType, &evt.HostGUID)
		if err != nil {
			log.Printf("Failed to parse HEAD event: %s\n", err)
		} else {
			parseSuccess = true
			agent.LastHeadEvent = evt
			agent.HostGUID = evt.HostGUID
		}

	default:
		if agent.EventCount > 10 {
			log.Printf("unhandled hone event type: %s\n", eventType)
		}
	}

	if parseSuccess {
		// event is chill
		//evt.Raw = lineBytes
		evt.Type = eventType
		agent.FillInEvent(evt)
		return evt
	}

	return nil
}

func parseInt(s string) int {
	res, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		log.Printf("failed to convert '%s' to int\n", s)
	}

	return int(res)
}

