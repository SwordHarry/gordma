// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import "errors"

type protectDomain struct {
	pd *C.struct_ibv_pd
}

func NewProtectDomain(ctx *rdmaContext) (*protectDomain, error) {
	pd, err := C.ibv_alloc_pd(ctx.ctx)
	if err != nil {
		return nil, err
	}
	return &protectDomain{
		pd: pd,
	}, err
}

func (p *protectDomain) Close() error {
	errno := C.ibv_dealloc_pd(p.pd)
	if errno != 0 {
		return errors.New("failed to dealloc PD")
	}
	p.pd = nil
	return nil
}
