// +build linux

package ibverbs

//#include <infiniband/verbs.h>
import "C"

type queuePair struct {

}

func NewQueuePair() (*queuePair, error) {
	return nil,nil
}
