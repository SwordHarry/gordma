// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import (
	"errors"
	"log"
	"math/rand"
	"time"
)

type queuePair struct {
	psn uint32
	qp  *C.struct_ibv_qp
	cq  *C.struct_ibv_cq
}

func NewQueuePair(pd *protectDomain, cq *completionQueue) (*queuePair, error) {
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
		psn: psn,
		qp:  qpC,
		cq:  cq.cq,
	}, nil
}

func (q *queuePair) Psn() uint32 {
	return q.psn
}

func (q *queuePair) Qpn() uint32  {
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
