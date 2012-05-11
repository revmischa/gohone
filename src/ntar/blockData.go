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
	"log"
	"unsafe"
)

const (
	PROCESS_EVENT_TYPE = 0x00000101
	CONNECTION_EVENT_TYPE = 0x00000102
	ENHANCED_PACKET_TYPE = 0x00000006
)

type HoneBlock C.hone_block
type ConnectionEvent C.connection_event_block
type ProcessEvent C.process_event_block
type PacketEvent struct {
	EventBlock *Block
}

type Event interface {}

// returns a ConnectionEvent, ProcessEvent or PacketEvent.
// Second return value is success
func (block *Block) BlockEvent() (Event, bool) {
	// try to get block type
	blockType, success := block.BlockType()
	if ! success {
		return nil, false
	}
	
	switch blockType {
	case PROCESS_EVENT_TYPE:	
		return block.ProcessEvent()
	case CONNECTION_EVENT_TYPE:
		return block.ConnectionEvent()
	case ENHANCED_PACKET_TYPE:
		return block.PacketEvent()
	default:
		log.Println("Got unknown blockType ", blockType)
	}

	return nil, false
}

func (block *Block) blockData() (*HoneBlock, bool) {
	var d HoneBlock
	res := int(C.ntar_get_block_data(block.Handle, (*unsafe.Pointer)(unsafe.Pointer(&d))))
	if res != 0 {
		log.Printf("get_block_data failed with res %d\n", res)
		return (*HoneBlock)(unsafe.Pointer(uintptr(0))), false
	}
	return &d, true
}

// connection event
func (block *Block) ConnectionEvent() (*ConnectionEvent, bool) {
	data, ok := block.blockData()
	if !ok {
		return (*ConnectionEvent)(unsafe.Pointer(uintptr(0))), ok
	}

	return (*ConnectionEvent)(unsafe.Pointer(&data)), true
}
func (data *ConnectionEvent) ConnectionID() uint {
	return uint(data.connection_id)
}
func (data *ConnectionEvent) ProcessID() uint {
	return uint(data.process_id)
}
func (data *ConnectionEvent) TimestampHigh() uint {
	return uint(data.timestamp_high)
}
func (data *ConnectionEvent) TimestampLow() uint {
	return uint(data.timestamp_low)
}

// process event
func (block *Block) ProcessEvent() (*ProcessEvent, bool) {
	data, ok := block.blockData()
	if !ok {
		return (*ProcessEvent)(unsafe.Pointer(uintptr(0))), ok
	}

	return (*ProcessEvent)(unsafe.Pointer(&data)), true
}
func (data *ProcessEvent) ProcessID() uint {
	return uint(data.process_id)
}
func (data *ProcessEvent) TimestampHigh() uint {
	return uint(data.timestamp_high)
}
func (data *ProcessEvent) TimestampLow() uint {
	return uint(data.timestamp_low)
}

// packet event
func (block *Block) PacketEvent() (*PacketEvent, bool) {
	evt := new(PacketEvent)
	evt.EventBlock = block

	return evt, true
}
func (evt *PacketEvent) Length() uint {
	data, ok := evt.EventBlock.blockData()
	if ! ok {
		return 0
	}
	
	return uint(data.Length)
}
/*func (data *ProcessEvent) TimestampHigh() uint {
	return uint(data.timestamp_high)
}
func (data *ProcessEvent) TimestampLow() uint {
	return uint(data.timestamp_low)
}
*/