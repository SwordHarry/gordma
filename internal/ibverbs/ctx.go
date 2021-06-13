// +build linux

package ibverbs

//#include <infiniband/verbs.h>
//#cgo linux LDFLAGS: -libverbs
//#include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"net"
	"unsafe"
)

type rdmaContext struct {
	Name string
	Port int
	Guid net.HardwareAddr
	ctx  *C.struct_ibv_context
}

func NewRdmaContext(name string, port,index int) (*rdmaContext, error) {
	var count C.int
	var ctx *C.struct_ibv_context
	var guid net.HardwareAddr
	deviceList, err := C.ibv_get_device_list(&count)
	if err != nil {
		return nil, err
	}
	if deviceList == nil || count == 0 {
		return nil, errors.New("failed to get devices list")
	}

	defer C.ibv_free_device_list(deviceList)
	devicePtr := deviceList
	device := *devicePtr
	for device != nil {
		ctx = C.ibv_open_device(device)
		if ctx != nil {
			var gid C.union_ibv_gid
			portC := C.uint8_t(port)
			indexC := C.int(index)
			errno, err := C.ibv_query_gid(ctx, portC, indexC, &gid)
			if errno != 0 || err != nil {
				return nil, err
			}
			guid = net.HardwareAddr(gid[8:])
			break
		}
		prevDevicePtr := uintptr(unsafe.Pointer(devicePtr))
		sizeofPtr := unsafe.Sizeof(devicePtr)
		devicePtr = (**C.struct_ibv_device)(unsafe.Pointer(prevDevicePtr + sizeofPtr))
		device = *devicePtr
	}
	if ctx == nil {
		return nil, fmt.Errorf("failed to open device %s", name)
	}
	return &rdmaContext{
		Name: name,
		ctx: ctx,
		Port: port,
		Guid: guid,
	}, nil
}

func (c *rdmaContext) Close() error  {
	errno := C.ibv_close_device(c.ctx)
	if errno != 0 {
		return errors.New("failed to close device")
	}
	c.ctx = nil
	return nil
}
