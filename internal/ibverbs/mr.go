// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import (
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"runtime"
	"unsafe"
)

type memoryRegion struct {
	remoteKey uint32
	isMmap    bool
	buf       []byte
	PD        *protectDomain
	mr        *C.struct_ibv_mr
}

func NewMemoryRegion(pd *protectDomain, size int, isMmap bool) (*memoryRegion, error) {
	var (
		buf []byte
		err error
	)
	if isMmap {
		// create buffer: why can not directly use make([]byte, size) ?
		const mrPort = unix.PROT_READ | unix.PROT_WRITE
		const mrFlags = unix.MAP_PRIVATE | unix.MAP_ANONYMOUS
		buf, err = unix.Mmap(-1, 0, size, mrPort, mrFlags)
		if err != nil {
			return nil, errors.New("mmap: failed to Mmap the buf")
		}
	} else {
		buf = make([]byte, size)
	}
	const access = IBV_ACCESS_LOCAL_WRITE | IBV_ACCESS_REMOTE_WRITE | IBV_ACCESS_REMOTE_READ
	mrC := C.ibv_reg_mr(pd.pd, unsafe.Pointer(&buf[0]), C.size_t(size), access)
	if mrC == nil {
		return nil, errors.New("ibv_reg_mr: failed to reg mr")
	}
	mr := &memoryRegion{
		buf: buf,
		PD:  pd,
		mr:  mrC,
		remoteKey: uint32(mrC.rkey),
	}
	runtime.SetFinalizer(mr, (*memoryRegion).finalize)
	return mr, nil
}

func (m *memoryRegion) RemoteKey() uint32 {
	return m.remoteKey
}

func (m *memoryRegion) String() string {
	if m.buf == nil {
		return "memoryRegion@closed"
	}
	return fmt.Sprintf("memoryRegion@%x[%d]", &m.buf[0], len(m.buf))
}

func (m *memoryRegion) finalize() {
	panic("finalized unclosed memory region")
}

func (m *memoryRegion) Close() error {
	errno := C.ibv_dereg_mr(m.mr)
	if errno != 0 {
		return errors.New("failed to dealloc mr")
	}
	if m.isMmap {
		err := unix.Munmap(m.buf)
		if err != nil {
			return err
		}
	} else {
		m.buf = nil
	}

	return nil
}
