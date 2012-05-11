package ntar

import (
	"hone"
	"ntar"
	"log"
)

type BlockChannel chan *ntar.Block

func NewReader() hone.Reader {
	return hone.Reader(new(Reader))
}

type Reader struct {
	Handle *ntar.Handle
}

func (reader *Reader) StartCapture(agent *hone.Agent) {
	done := false
	for (! done) {
		section := reader.Handle.GetNextSection()
		for section != nil {
			log.Println("Got section")
			
			block := section.GetNextBlock()
			for block != nil {
				// parse block
				
				// this cannot be asynchronous because ntar expects
				// the current block to be closed before retrieving
				// the next block
				reader.parseBlock(block, agent)

				// retrieve next block
				block = section.GetNextBlock()
			}

			section.Destroy()
			section = reader.Handle.GetNextSection()
		}
		
		// error
		//reader.Handle.Open()
		log.Println("Done")
		done = true
	}
}

func (reader *Reader) parseBlock(block *ntar.Block, agent *hone.Agent) {
	defer block.Destroy()
	
	data, success := block.BlockEvent()
	if ! success {
		return
	}

	//defer data.Destroy()

	switch event := data.(type) {
	case *ntar.ProcessEvent:
		log.Println("Process event")
	case *ntar.PacketEvent:
		log.Println("Got packet event: ", event.ConnectionID())
	case *ntar.ConnectionEvent:
		log.Println("Got connection event", event.ProcessID(), event.ConnectionID())
	default:
		log.Println("Got unknown event: ", event)
	}
}

func (reader *Reader) OpenCaptureFile() {
	handle := ntar.NewHandle()
	reader.Handle = handle
	handle.Open()
}

func (reader *Reader) CloseCaptureFile() {
	reader.Handle.Close()
}