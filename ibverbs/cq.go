// +build linux

package ibverbs

//#include <infiniband/verbs.h>
////#include <rdma/rdma_cma.h>
import "C"
import (
	"errors"
	"golang.org/x/sys/unix"
)

type completionQueue struct {
	cq      *C.struct_ibv_cq
	channel *C.struct_ibv_comp_channel
}

func NewCompletionQueue(ctx *rdmaContext, cqe int) (*completionQueue, error) {
	compChannel, err := C.ibv_create_comp_channel(ctx.ctx)
	if compChannel == nil {
		return nil, errors.New("failed to create compChannel")
	}
	if err != nil {
		return nil, err
	}
	if err := unix.SetNonblock(int(compChannel.fd), true); err != nil {
		return nil, err
	}
	// TODO: err: protocol not supported? but the cq can be created
	cq, err := C.ibv_create_cq(ctx.ctx, 10, nil, compChannel, 0)
	if cq != nil {
		return &completionQueue{
			cq:      cq,
			channel: compChannel,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, errors.New("unknwon error")
}

func (c *completionQueue) Close() error {
	channel := c.cq.channel
	errno := C.ibv_destroy_cq(c.cq)
	if errno != 0 {
		return errors.New("ibv_destroy_cq failed")
	}
	if channel != nil {
		errno := C.ibv_destroy_comp_channel(channel)
		if errno != 0 {
			return errors.New("ibv_destroy_comp_channel failed")
		}
	}
	return nil
}
