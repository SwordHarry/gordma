// hello.go
package main

import "C"
import (
	"fmt"
	"gordma/internal/ibverbs"
)

func main() {
	c, err := ibverbs.NewContext("rxe_0", 1, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println(c)
	pd := ibverbs.NewProtectDomain(c)
	fmt.Println(pd)
	fmt.Println(pd.Close())
}
