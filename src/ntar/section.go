package ntar

// #cgo CFLAGS: -I/home/bobo/dev/ntar
// #cgo LDFLAGS: -lntar
// #include <stdlib.h>
// #include <ntar.h>
import "C"

import (
	"log"
	"unsafe"
)

type SectionHandle **C.ntar_section_handle

type Section struct {
	Handle SectionHandle
}

func NewSection() *Section {
	section := new(Section)

	handle := new(*C.ntar_section_handle)
	section.Handle = handle

	return section
}

func (section *Section) GetNextBlock() *Block {
	block := NewBlock()

	ret := C.int(C.ntar_get_next_block(*section.Handle, block.Handle))
	if ret != C.int(C.NTAR_SUCCESS) {
		log.Printf("got next block failed = %d\n", ret)
		return nil
	}

	return block
}

func (section *Section) Destroy() {
	if section.Handle == nil {
		return
	}

	C.ntar_close_section(*section.Handle)
	C.free(unsafe.Pointer(section.Handle))
}
