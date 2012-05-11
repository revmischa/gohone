package ntar

/*
 #cgo CFLAGS: -I/home/bobo/dev/ntar
 #cgo LDFLAGS: -lntar

 #include <ntar.h>
 #include <stdlib.h>
*/
import "C"

import (
	"log"
	"unsafe"
)

type BlockOptions struct {
	OptionRoot NtarOption
}

func (opts *BlockOptions) Dump() {
	head := opts.OptionRoot
	for unsafe.Pointer(&head) != nil {
		log.Printf("Option code: %#v, data: %#v\n", uint(head.code), head.data)

		next := head.next
		if unsafe.Pointer(next) == nil {
			return
		}

		// there has got to be a less grimy way to traverse a linked
		// list than this
		nextVal := (*NtarOption)(unsafe.Pointer(next))
		head = *nextVal
	}
}

func (opts *BlockOptions) Find(opt uint) unsafe.Pointer {
	head := opts.OptionRoot
	for unsafe.Pointer(&head) != nil {
		if uint(head.code) == opt {
			return unsafe.Pointer(head.data)
		}

		next := head.next
		if unsafe.Pointer(next) == nil {
			return nil
		}

		// there has got to be a less grimy way to traverse a linked
		// list than this
		nextVal := (*NtarOption)(unsafe.Pointer(next))
		head = *nextVal
	}

	return nil
}	

