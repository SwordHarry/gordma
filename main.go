// hello.go
package main

import "C"
import (
	"fmt"
	"gordma/internal/ibverbs"
)

func main() {
	c, err := ibverbs.NewRdmaContext("rxe_0", 1, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	pd := ibverbs.NewProtectDomain(c)
	fmt.Println("pd", pd)
	mr, err := ibverbs.NewMemoryRegion(pd, 1024, false)
	if err != nil {
		panic(err)
	}
	fmt.Println(mr, mr.RemoteKey())

	// ---------------- close ---------------
	fmt.Println(mr.Close())
	fmt.Println(pd.Close())
	fmt.Println(c.Close())
}
