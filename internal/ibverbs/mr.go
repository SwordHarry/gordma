// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"
import "fmt"

type MemoryRegion struct {
	buf []byte
	PD  *ProtectDomain
	mr  *C.struct_ibv_mr
}


func NewMemoryRegion(pd *ProtectDomain)  {

}

func (m *MemoryRegion) String() string {
	if m.buf == nil {
		return "MemoryRegion@closed"
	}
	return fmt.Sprintf("MemoryRegion@%x[%d]", &m.buf[0], len(m.buf))
}
