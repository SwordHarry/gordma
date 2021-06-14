// +build linux

package ibverbs

//#include <infiniband/verbs.h>
//#cgo linux LDFLAGS: -libverbs
//#include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"unsafe"
)

type rdmaContext struct {
	Name string
	Port int
	Guid net.HardwareAddr
	ctx  *C.struct_ibv_context
}

type rlimir struct {
	cur uint64
	max uint64
}

func init() {
	// skip memlock check if there is no IB hardware
	var r rlimir
	_, _, err := unix.Syscall(unix.SYS_GETRLIMIT, unix.RLIMIT_MEMLOCK, uintptr(unsafe.Pointer(&r)), 0)
	if err != 0 {
		panic(err.Error())
	}
	fmt.Println(r)
	//const maxUint64 = 1<<64 - 1
	//if r.cur != uint64(maxUint64) || r.max != uint64(maxUint64) {
	//	panic("ib: memlock rlimit is not unlimited")
	//}
}

func NewRdmaContext(name string, port, index int) (*rdmaContext, error) {
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
	for device != nil && ctx == nil {
		ctx = C.ibv_open_device(device)
		var gid C.union_ibv_gid
		portC := C.uint8_t(port)
		indexC := C.int(index)
		errno, err := C.ibv_query_gid(ctx, portC, indexC, &gid)
		if errno != 0 || err != nil {
			return nil, err
		}
		guid = net.HardwareAddr(gid[8:])
		// next device
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
		ctx:  ctx,
		Port: port,
		Guid: guid,
	}, nil
}

func (c *rdmaContext) Close() error {
	errno := C.ibv_close_device(c.ctx)
	if errno != 0 {
		return errors.New("failed to close device")
	}
	c.ctx = nil
	return nil
}

func (c *rdmaContext) String() string {
	return fmt.Sprintf("rdmaContext: \n name: %s\n port: %d\n guid: %s\n ", c.Name, c.Port, c.Guid)
}
