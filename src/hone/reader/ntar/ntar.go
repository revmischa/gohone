package ntar

import (
	"hone"
	"ntar"
	"log"
)

const (
	PROCESS_EVENT_TYPE = 0x00000101
	CONNECTION_EVENT_TYPE = 0x00000102
	ENHANCED_PACKET_TYPE = 0x00000006
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
	
	// try to get block type
	blockType, success := block.BlockType()
	if ! success {
		return
	}

	/*
	blockData, success := block.BlockData()
	if ! success {
		return
	}
	log.Printf("blockdata.len=%#v\n", blockData.Length)
	*/
		
	switch (blockType) {
	case PROCESS_EVENT_TYPE:	
		log.Println("Got process event")
	case CONNECTION_EVENT_TYPE:
		log.Println("Got connection event")
	case ENHANCED_PACKET_TYPE:
		log.Println("Got packet event")
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