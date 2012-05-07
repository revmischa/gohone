package ntar

/*
 #cgo CFLAGS: -I/home/bobo/dev/ntar
 #cgo LDFLAGS: -lntar
 
 #include <ntar.h>
 #include <stdlib.h>

 struct _hone_block {
 int Length;
 };
 typedef struct _hone_block hone_block;
 
*/
import "C"

import (
	"unsafe"
	"log"
)

type BlockHandle *C.ntar_block_handle
type HoneBlock C.hone_block

type Block struct {
	Handle BlockHandle
}

func NewBlock() *Block {
	block := new(Block)

	var handle BlockHandle
	block.Handle = handle

	return block
}

func (block *Block) BlockType() (int, bool) {
	var t C.u_int32
	res := int(C.ntar_get_block_type(block.Handle, &t))
	if res != 0 {
		log.Printf("get_block_type failed with res %d\n", res)
		return 0, false
	}
	return int(t), true
}

func (block *Block) BlockData() (*HoneBlock, bool) {
	var d HoneBlock
	res := int(C.ntar_get_block_data(block.Handle, (*unsafe.Pointer)(unsafe.Pointer(&d))))
	if res != 0 {
		log.Printf("get_block_data failed with res %d\n", res)
		return (*HoneBlock)(unsafe.Pointer(uintptr(0))), false
	}
	return &d, true
}

func (block *Block) Destroy() {
	if block.Handle == nil {
		return
	}

	C.ntar_close_block(block.Handle)

	block.Handle = nil
}
