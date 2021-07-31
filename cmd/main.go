package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -lobjc -framework Foundation -framework CoreFoundation -framework AppKit

#import <DetectorAppDelegate.h>
*/
import "C"
import (
	"fmt"
	"github.com/progrium/macdriver/objc"
	"log"
	"runtime"
	"unsafe"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	runtime.LockOSThread()

	application := objc.Get("NSApplication").Send("sharedApplication")
	applicationDelegate := objc.Get("DetectorAppDelegate").Alloc().Init()
	application.Send("setDelegate:", applicationDelegate)
	application.Send("run")
}

//export queryWire
func queryWire(pathPointer unsafe.Pointer, createdAtPointer unsafe.Pointer) {
	if pathPointer == nil || createdAtPointer == nil {
		return
	}

	path := objc.ObjectPtr(uintptr(pathPointer))
	path.Retain()
	createdAt := objc.ObjectPtr(uintptr(createdAtPointer))
	createdAt.Retain()

	fmt.Println(path, createdAt)
}
