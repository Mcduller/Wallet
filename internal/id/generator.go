package id

import (
	"fmt"
	"sync/atomic"
)

type Generator struct {
	counter atomic.Uint64
}

var _ IDGenerator = (*Generator)(nil)

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) NextWalletID() string {
	return fmt.Sprintf("w_%d", g.counter.Add(1))
}
