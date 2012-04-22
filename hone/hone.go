package hone

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
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
	Type   CaptureEventType
	Packet []byte

	HostGUID string

	ConnectionState rune
	Direction       rune
	Sockfd          uint64
	Proto           string
	Src             string
	Dst             string
	Len             uint64

	PID  int
	PPID int
	TGID int
	UID  int
	GID  int

	Executable string
	Argv       string
}

// capture event type

// class that reads and dispatches capture events
type Agent struct {
	CaptureFile *os.File
	Stopped     bool
	EventCount  uint64

	ServerAddress     string
	ServerPort        uint
	ServerConn        net.Conn
	ConnectedToServer bool
	Connecting        bool

	// mapping of sockfd -> last sock event
	SockEvents map[uint64]*CaptureEvent
	// mapping of tgid -> last exec event
	ExecEvents map[int]*CaptureEvent

	// pid of thread communicating with collector
	CommunicationPID int
}

func NewAgent(serverAddr string, serverPort uint) *Agent {
	agent := new(Agent)

	// initialization
	agent.SockEvents = make(map[uint64]*CaptureEvent)
	agent.ExecEvents = make(map[int]*CaptureEvent)

	agent.ServerAddress = serverAddr
	agent.ServerPort = serverPort

	return agent
}

func (agent *Agent) Connect() {
	if agent.Connecting {
		return
	}

	// try connecting to collection server
	server := agent.ServerAddress + ":" + strconv.FormatUint(uint64(agent.ServerPort), 10)

	agent.CommunicationPID = os.Getpid()

	agent.Connecting = true
	conn, err := net.Dial("tcp", server)

	if err != nil {
		log.Printf(fmt.Sprintf("Failed to connect to %s: %s\n", server, err))
		agent.ConnectedToServer = false
		agent.Connecting = false
		return
	}

	agent.ServerConn = conn
	agent.ConnectedToServer = true
	agent.Connecting = false
}

// open kernel hone event module
func (agent *Agent) OpenCaptureFile() {
	var err error
	agent.CaptureFile, err = os.Open(capFilePath)

	// TODO: check if module is loaded

	if err != nil {
		log.Fatalf("Error opening capture file for reading: %s\n", err)
	}
}

func (agent *Agent) CloseCaptureFile() {
	agent.CaptureFile.Close()
}

// opens capture, starts sending events to event channel
func (agent *Agent) Run() EventChan {
	// open capture
	agent.OpenCaptureFile()

	// create a channel to receive capture events
	ec := make(EventChan)

	// start reading file
	go func() {
		for !agent.Stopped {
			reader := bufio.NewReader(agent.CaptureFile)
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

			agent.EventCount++

			// parse input
			evt := agent.ParseHoneEventLine(line)
			if evt == nil {
				continue
			}

			// success
			ec <- evt
		}

		agent.CloseCaptureFile()
	}()

	return ec
}

func (agent *Agent) Stop() {
	agent.Stopped = true
}

// parses a line from /dev/honet into a CaptureEvent
func (agent *Agent) ParseHoneEventLine(lineBytes []byte) *CaptureEvent {
	line := string(lineBytes)

	parseSuccess := false

	//log.Printf("line: %s\n", line)

	// parse timestamp and event type
	evt := new(CaptureEvent)
	var delta float64
	var eventType CaptureEventType
	parsed, err := fmt.Sscanf(line, "%f %s", &delta, &eventType)
	if err == nil && parsed == 2 {
		// handle event types
		//procSpec := "%d %d %d %d %d"
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
				evt.Argv = matches[7]

				agent.ExecEvents[evt.TGID] = evt
			}

		case "PAKT":
			// packet
			re := regexp.MustCompile("([IO]) ([A-Fa-f0-9]+) (\\S+) (\\S+) -> (\\S+) (\\d+)")
			matches := re.FindStringSubmatch(line)

			if len(matches) == 7 {
				evt.Direction = rune(matches[1][0])
				evt.Sockfd, _ = strconv.ParseUint(matches[2], 16, 0)
				evt.Proto = matches[3]
				evt.Src = matches[4]
				evt.Dst = matches[5]

				evtlen, err := strconv.ParseUint(matches[6], 10, 0)
				if err == nil {
					evt.Len = evtlen
				} else {
					fmt.Printf("Failed to parse length %s: %s\n", matches[6], err)
				}

				// attempt to locate corresponding info for this socket
				// find last sock event of matching sockfd
				sockEvt := agent.SockEvents[evt.Sockfd]
				if sockEvt != nil {
					evt.PID = sockEvt.PID
					evt.PPID = sockEvt.PPID
					evt.TGID = sockEvt.TGID
					evt.UID = sockEvt.UID
					evt.GID = sockEvt.GID
				} else {
					// fmt.Printf("Failed to find PID for sockFD %d\n\n", evt.Sockfd)
					// we're gonna ignore this, because we have no
					// mapping for the process yet
					return nil
				}

				parseSuccess = true
			}

		case "SOCK":
			// socket
			re := regexp.MustCompile("([CO]) " + procSpec + " ([A-Fa-f0-9]+)")
			matches := re.FindStringSubmatch(line)

			if len(matches) == 8 {
				evt.Direction = rune(matches[1][0])
				evt.Sockfd, _ = strconv.ParseUint(matches[7], 16, 0)

				evt.PID = parseInt(matches[2])
				evt.PPID = parseInt(matches[3])
				evt.TGID = parseInt(matches[4])
				evt.UID = parseInt(matches[5])
				evt.GID = parseInt(matches[6])

				if evt.PID != 0 && evt.Sockfd != 0 {
					parseSuccess = true
					agent.SockEvents[evt.Sockfd] = evt
				}
			}

		case "HEAD":
			// capture header, host GUID
			_, err := fmt.Sscanf(line, "%f %s %s", &delta, &eventType, &evt.HostGUID)
			if err != nil {
				log.Printf("Failed to parse HEAD event: %s\n", err)
			} else {
				parseSuccess = true
			}

		default:
			if agent.EventCount > 10 {
				log.Printf("unhandled hone event type: %s\n", eventType)
			}
		}
	}

	if !parseSuccess {
		if agent.EventCount > 10 {
			log.Printf("Failed to parse line '%s': %s\n", line, err)
		}
	}

	if parseSuccess {
		// event is chill
		//evt.Packet = lineBytes
		evt.Type = eventType
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
