package ntar

import (
	"hone"
	"ntar"
	"log"
)

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
				log.Printf("Got block\n")
				
				block.Destroy()
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

func (reader *Reader) OpenCaptureFile() {
	handle := ntar.NewHandle()
	reader.Handle = handle
	handle.Open()
}

func (reader *Reader) CloseCaptureFile() {
	reader.Handle.CloseHandle()
}