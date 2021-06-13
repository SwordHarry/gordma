// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import "errors"

type ProtectDomain struct {
	pd *C.struct_ibv_pd
}

func NewProtectDomain(ctx *Context) *ProtectDomain {
	return &ProtectDomain{
		pd: C.ibv_alloc_pd(ctx.ctx),
	}
}

func (p *ProtectDomain) Close() error {
	errno := C.ibv_dealloc_pd(p.pd)
	if errno != 0 {
		return errors.New("failed to dealloc PD")
	}
	p.pd = nil
	return nil
}
