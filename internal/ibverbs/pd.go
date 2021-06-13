// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import "errors"

type protectDomain struct {
	pd *C.struct_ibv_pd
}

func NewProtectDomain(ctx *rdmaContext) *protectDomain {
	return &protectDomain{
		pd: C.ibv_alloc_pd(ctx.ctx),
	}
}

func (p *protectDomain) Close() error {
	errno := C.ibv_dealloc_pd(p.pd)
	if errno != 0 {
		return errors.New("failed to dealloc PD")
	}
	p.pd = nil
	return nil
}
