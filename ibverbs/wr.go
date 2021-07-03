// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import "unsafe"

type sendWorkRequest struct {
	mr        *memoryRegion
	sendWr    *C.struct_ibv_send_wr
}

type receiveWorkRequest struct {
	mr        *memoryRegion
	recvWr    *C.struct_ibv_recv_wr
}

func NewSendWorkRequest(mr *memoryRegion) *sendWorkRequest {
	var sendWr C.struct_ibv_send_wr
	return &sendWorkRequest{
		mr:     mr,
		sendWr: &sendWr,
	}
}

func NewReceiveWorkRequest(mr *memoryRegion) *receiveWorkRequest  {
	var recvWr C.struct_ibv_recv_wr
	return &receiveWorkRequest{
		mr:     mr,
		recvWr: &recvWr,
	}
}

func (s *sendWorkRequest) createWrId() C.uint64_t {
	return C.uint64_t(uintptr(unsafe.Pointer(&(s.sendWr))))
}

func (r *receiveWorkRequest) createWrId() C.uint64_t  {
	return C.uint64_t(uintptr(unsafe.Pointer(&(r.recvWr))))
}
