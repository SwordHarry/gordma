// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import (
	"errors"
	"golang.org/x/sys/unix"
	"log"
)

type completionQueue struct {
	cqe     int
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
	cq, err := C.ibv_create_cq(ctx.ctx, C.int(cqe), nil, compChannel, 0)
	if cq == nil {
		if err != nil {
			log.Println("cq", err)
			return nil, err
		}
		return nil, errors.New("unknown error")
	}
	return &completionQueue{
		cqe:     cqe,
		cq:      cq,
		channel: compChannel,
	}, nil
}

func (c *completionQueue) Cqe() int {
	return c.cqe
}

func (c *completionQueue) Close() error {
	channel := c.cq.channel
	errno := destroyCQ(c.cq)
	if errno != 0 {
		return errors.New("ibv_destroy_cq failed")
	}
	if channel != nil {
		errno := destroyCompChannel(channel)
		if errno != 0 {
			return errors.New("ibv_destroy_comp_channel failed")
		}
	}
	return nil
}

func destroyCQ(cq *C.struct_ibv_cq) C.int {
	return C.ibv_destroy_cq(cq)
}

func destroyCompChannel(channel *C.struct_ibv_comp_channel) C.int {
	return C.ibv_destroy_comp_channel(channel)
}
