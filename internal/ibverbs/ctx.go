// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type Context struct {
	Name string
	ctx  *C.struct_ibv_context
}

func NewContext(name string) (*Context, error) {
	var count C.int
	var ctx *C.struct_ibv_context
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
		if device.name == name {
			ctx, err = C.ibv_open_device(device)
			if err != nil  {
				return nil, err
			}
			if ctx == nil {
				return nil, fmt.Errorf("failed to open device %s", name)
			}
		}
		prevDevicePtr := uintptr(unsafe.Pointer(devicePtr))
		sizeofPtr := unsafe.Sizeof(devicePtr)
		devicePtr = (**C.struct_ibv_device)(prevDevicePtr + sizeofPtr)
		device = *devicePtr
	}

	return &Context{
		Name: name,
		ctx: ctx,
	}, nil
}
