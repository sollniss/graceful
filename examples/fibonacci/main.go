package main

import (
	"context"
	"errors"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/sollniss/graceful"
)

var ErrInterrupted = errors.New("received interrupt signal")

type fibcalc struct {
	stopped bool
}

func fib(n int) *big.Int {
	var f1 *big.Int = big.NewInt(1)
	var f2 *big.Int = big.NewInt(0)
	for i := 0; i < n; i++ {
		time.Sleep(100 * time.Millisecond)
		f1, f2 = f2, f1.Add(f1, f2)
	}
	return f2
}

func (c *fibcalc) BeginCalculating() error {
	var i int = 1
	for ; ; i++ {
		res := fib(i)
		log.Printf("F_%d: %d", i, res)
		if c.stopped {
			break
		}
		if i == math.MaxInt {
			return errors.New("loop variable overflowed")
		}
	}

	return ErrInterrupted
}

func (c *fibcalc) StopCalculating() {
	c.stopped = true
}

func main() {

	fib := fibcalc{}

	ctx := graceful.NotifyShutdown()
	start := func() error {
		return fib.BeginCalculating()
	}

	shutdown := func(_ context.Context) error {
		fib.StopCalculating()
		return nil
	}

	// Since we don't handle the context in shutdown, the timeout value is not used.
	err := graceful.Start(ctx, start, ErrInterrupted, shutdown, 0)
	if err != nil {
		log.Printf("error during shutdown: %s", err)
		return
	}
	log.Print("gracefully shut down")
}
