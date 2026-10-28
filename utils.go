package hl7converter

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type RetryableOnce struct {
	mu sync.Mutex

	done atomic.Bool
	err  error // storage last call err
}

const maxRetries = 3

func (o *RetryableOnce) Do(fn func() error) error {
	if o.done.Load() {
		return nil
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	if o.done.Load() { // repeated check
		return nil
	}

	var err error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err = fn()
		if err == nil {
			o.done.Store(true)
			o.err = nil

			return nil
		}

		time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)
	}

	o.err = fmt.Errorf("failed after %d attempts: %w", maxRetries, err)

	return o.err
}
