package debouncer

import (
	"context"
	"sync"
	"time"
)

type Debouncer struct {
	delay   time.Duration
	input   chan struct{}
	output  chan struct{}
	mu      sync.Mutex
	running bool
}

// New creates a debouncer with specified delay
// Typical value: 300ms
func New(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay:   delay,
		input:   make(chan struct{}, 1),
		output:  make(chan struct{}),
	}
}

// Start begins debounce processing
// When events arrive via Trigger(), waits 'delay' before sending signal
// If more events arrive during delay, timer resets
func (d *Debouncer) Start(ctx context.Context) {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return
	}
	d.running = true
	d.mu.Unlock()

	go func() {
		defer func() {
			d.mu.Lock()
			d.running = false
			d.mu.Unlock()
		}()

		var timer *time.Timer
		var timerC <-chan time.Time

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				// Don't close output channel on cancellation
				// Let it be closed when the debouncer is naturally done
				return

			case <-d.input:
				// Reset timer
				if timer != nil {
					timer.Stop()
				}
				timer = time.NewTimer(d.delay)
				timerC = timer.C

			case <-timerC:
				// Timer fired, send output signal
				select {
				case d.output <- struct{}{}:
				case <-ctx.Done():
					if timer != nil {
						timer.Stop()
					}
					return
				}
				// Clear timer until next trigger
				timer = nil
				timerC = nil
			}
		}
	}()
}

// Trigger sends an event to debouncer (non-blocking)
// Safe to call concurrently
func (d *Debouncer) Trigger() {
	select {
	case d.input <- struct{}{}:
		// Trigger sent successfully
	default:
		// Input channel full, trigger already pending
	}
}

// Output returns channel that emits after quiet period
func (d *Debouncer) Output() <-chan struct{} {
	return d.output
}
