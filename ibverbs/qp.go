// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import (
	"errors"
	"gordma/common"
	"log"
	"math/rand"
	"time"
	"unsafe"
)

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

// queuePair QP
type queuePair struct {
	psn  uint32
	port int
	qp   *C.struct_ibv_qp
	cq   *C.struct_ibv_cq
}

func NewQueuePair(ctx *rdmaContext, pd *protectDomain, cq *completionQueue) (*queuePair, error) {
	initAttr := C.struct_ibv_qp_init_attr{}
	initAttr.send_cq = cq.cq
	initAttr.recv_cq = cq.cq
	cqe := cq.Cqe()
	initAttr.cap.max_send_wr = C.uint32_t(cqe)
	initAttr.cap.max_recv_wr = C.uint32_t(cqe)
	initAttr.cap.max_send_sge = 1
	initAttr.cap.max_recv_sge = 1
	//initAttr.cap.max_inline_data = 64
	initAttr.qp_type = IBV_QPT_RC
	// make everything signaled. avoids the problem with inline
	// sends filling up the send queue of the cq
	initAttr.sq_sig_all = 1

	qpC, err := C.ibv_create_qp(pd.pd, &initAttr)
	if qpC == nil {
		if err != nil {
			log.Println("qp", err)
			return nil, err
		}
		return nil, errors.New("qp: unknown error")
	}

	// create psn
	psn := rand.New(rand.NewSource(time.Now().UnixNano())).Uint32() & 0xffffff
	return &queuePair{
		psn:  psn,
		port: ctx.Port,
		qp:   qpC,
		cq:   cq.cq,
	}, nil
}

func (q *queuePair) Psn() uint32 {
	return q.psn
}

func (q *queuePair) Qpn() uint32 {
	return uint32(q.qp.qp_num)
}

func (q *queuePair) Close() error {
	if q.qp == nil {
		return nil
	}

	errno := C.ibv_destroy_qp(q.qp)
	if errno != 0 {
		return errors.New("ibv_destroy_qp failed")
	}
	q.qp = nil
	return nil
}

func (q *queuePair) modify(attr *C.struct_ibv_qp_attr, mask int) error {
	errno := C.ibv_modify_qp(q.qp, attr, mask)
	return common.NewErrorOrNil("ibv_modify_qp",errno)
}

func (q *queuePair) Init() error {
	attr := C.struct_ibv_qp_attr{}
	attr.qp_state = C.IBV_QPS_INIT
	attr.pkey_index = 0
	attr.port_num = C.uint8_t(q.port)
	// allow RDMA write
	attr.qp_access_flags = IBV_ACCESS_LOCAL_WRITE | IBV_ACCESS_REMOTE_READ | IBV_ACCESS_REMOTE_WRITE
	mask := C.IBV_QP_STATE | C.IBV_QP_PKEY_INDEX | C.IBV_QP_PORT | C.IBV_QP_ACCESS_FLAGS
	return q.modify(&attr, mask)
}

// Ready2Receive RTR
// 尽管是 RTR，但是 send 和 receive 的配置都在这里提前配好
func (q *queuePair) Ready2Receive(destGid uint16, destQpn, destPsn uint32) error {
	attr := C.struct_ibv_qp_attr{}
	attr.qp_state = C.IBV_QPS_RTR
	attr.path_mtu = C.IBV_MTU_2048
	attr.dest_qp_num = C.uint32_t(destQpn)
	attr.rq_psn = C.uint32_t(destPsn)
	// this must be > 0 to avoid IBV_WC_REM_INV_REQ_ERR
	attr.max_dest_rd_atomic = 1
	// Minimum RNR NAK timer (range 0..31)
	attr.min_rnr_timer = 26
	attr.ah_attr.is_global = 0
	attr.ah_attr.dgid = C.uint16_t(destGid)
	//  attr.ah_attr.dlid = C.uint16_t(destLid)
	attr.ah_attr.sl = 0
	attr.ah_attr.src_path_bits = 0
	attr.ah_attr.port_num = C.uint8_t(q.port)
	mask := C.IBV_QP_STATE | C.IBV_QP_AV | C.IBV_QP_PATH_MTU | C.IBV_QP_DEST_QPN |
		C.IBV_QP_RQ_PSN | C.IBV_QP_MAX_DEST_RD_ATOMIC | C.IBV_QP_MIN_RNR_TIMER
	return q.modify(&attr, mask)
}

// Ready2Send RTS
func (q *queuePair)  Ready2Send() error  {
	attr := C.struct_ibv_qp_attr{}
	attr.qp_state = C.IBV_QPS_RTS
	// Local ack timeout for primary path.
	// Timeout is calculated as 4.096e-6*(2**attr.timeout) seconds.
	attr.timeout = 14
	// Retry count (7 means forever)
	attr.retry_cnt = 6
	// RNR retry (7 means forever)
	attr.rnr_retry = 6
	attr.sq_psn = C.uint32_t(q.psn)
	// this must be > 0 to avoid IBV_WC_REM_INV_REQ_ERR
	attr.max_rd_atomic = 1
	mask := C.IBV_QP_STATE | C.IBV_QP_TIMEOUT | C.IBV_QP_RETRY_CNT | C.IBV_QP_RNR_RETRY |
		C.IBV_QP_SQ_PSN | C.IBV_QP_MAX_QP_RD_ATOMIC
	return q.modify(&attr, mask)
}

func (q  *queuePair) PostSend(wr *sendWorkRequest) error {
	return q.PostSendImm(wr,0)
}
func (q *queuePair) PostSendImm(wr *sendWorkRequest,  imm uint32) error {
	if imm > 0 {
		// post_send_immediately
	} else {
		// post_send
		wr.sendWr.opcode = IBV_WR_SEND
		wr.sendWr.send_flags = IBV_SEND_SIGNALED
	}

	if wr.mr != nil {

	}  else  {

	}
	wr.sendWr.wr_id = C.uint64_t(uintptr(unsafe.Pointer(&(wr.sendWr))))
	var bad *C.struct_ibv_send_wr
	errno := C.ibv_post_send(q.qp, &wr.sendWr, &bad)
	return common.NewErrorOrNil("ibv_post_send", errno)
}
