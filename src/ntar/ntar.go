package ntar

// #cgo CFLAGS: -I/home/bobo/dev/ntar
// #cgo LDFLAGS: -lntar
/*
  #include <stdlib.h>
  #include <ntar.h>
*/
import "C"
import (
	"log"
	"unsafe"
)

const (
	capFilePath = "/dev/hone"
)

type FileHandle **C.ntar_file_handle

type Handle struct {
	CaptureFile FileHandle
}

func NewHandle() *Handle {
	h := new(Handle)
	return h
}

func (handle *Handle) GetNextSection() *Section {
	section := NewSection()

	ret := C.int(C.ntar_get_next_section(*handle.CaptureFile, section.Handle))
	if ret != C.int(C.NTAR_SUCCESS) {
		log.Printf("got next section failed = %d\n", ret)
		return nil
	}

	return section
}

func (handle *Handle) CloseHandle() {
	if handle.CaptureFile == nil {
		return
	}

	C.ntar_close(*handle.CaptureFile)
	//	C.free(unsafe.Pointer(*handle.CaptureFile))

	handle.CaptureFile = nil
}

func (handle *Handle) Open() int {
	handle.CloseHandle()
	fh := new(*C.ntar_file_handle)
	handle.CaptureFile = fh

	fileNameC := C.CString(capFilePath)
	defer C.free(unsafe.Pointer(fileNameC))
	flagsC := C.CString("r")
	defer C.free(unsafe.Pointer(flagsC))

	ret := int(C.ntar_open(fileNameC, flagsC, fh))
	if ret != C.NTAR_SUCCESS {
		log.Fatalf("Failed to open %s, ret=%d\n", capFilePath, ret)
	}

	return ret
}
