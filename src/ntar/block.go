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

type BlockHandle **C.ntar_block_handle

type Block struct {
	Handle BlockHandle
}

func NewBlock() *Block {
	block := new(Block)

	handle := new(*C.ntar_block_handle)
	block.Handle = handle

	return block
}

func (block *Block) Destroy() {
	if block.Handle == nil {
		return
	}

	log.Printf("Closing block handle %#v\n", block.Handle)
	C.ntar_close_block(unsafe.Pointer(*block.Handle))
	C.free(unsafe.Pointer(block.Handle))

	block.Handle = nil
}
