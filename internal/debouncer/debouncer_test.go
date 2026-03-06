package debouncer

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebouncer_SingleEvent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	debouncer := New(300 * time.Millisecond)
	debouncer.Start(ctx)

	start := time.Now()
	debouncer.Trigger()

	select {
	case <-debouncer.Output():
		duration := time.Since(start)
		// Should fire after ~300ms (±50ms tolerance)
		assert.InDelta(t, 300, duration.Milliseconds(), 50)
	case <-time.After(1 * time.Second):
		t.Fatal("debouncer did not fire")
	}
}

func TestDebouncer_RapidEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	debouncer := New(300 * time.Millisecond)
	debouncer.Start(ctx)

	// Trigger 10 events rapidly
	for i := 0; i < 10; i++ {
		debouncer.Trigger()
		time.Sleep(50 * time.Millisecond) // 50ms between events
	}

	// Should only get 1 output
	outputs := 0
	timeout := time.After(1 * time.Second)

	for {
		select {
		case <-debouncer.Output():
			outputs++
		case <-timeout:
			assert.Equal(t, 1, outputs, "expected exactly 1 output")
			return
		}
	}
}

func TestDebouncer_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	debouncer := New(300 * time.Millisecond)
	debouncer.Start(ctx)

	// Cancel immediately without triggering
	cancel()

	// Give some time for cancellation to process
	time.Sleep(50 * time.Millisecond)

	// Should not receive output
	select {
	case <-debouncer.Output():
		t.Fatal("received output after cancellation")
	case <-time.After(100 * time.Millisecond):
		// Expected: no output
	}
}

func TestDebouncer_MultipleStart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	debouncer := New(100 * time.Millisecond)
	
	// Starting multiple times should not cause issues
	debouncer.Start(ctx)
	debouncer.Start(ctx)
	debouncer.Start(ctx)

	debouncer.Trigger()

	// Should still work normally
	select {
	case <-debouncer.Output():
		// Expected
	case <-time.After(500 * time.Millisecond):
		t.Fatal("debouncer did not fire")
	}
}

func TestDebouncer_TriggerNonBlocking(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	debouncer := New(300 * time.Millisecond)
	debouncer.Start(ctx)

	// Trigger multiple times rapidly without blocking
	for i := 0; i < 100; i++ {
		debouncer.Trigger()
	}

	// Should still only get one output
	select {
	case <-debouncer.Output():
		// Expected first output
	case <-time.After(1 * time.Second):
		t.Fatal("debouncer did not fire")
	}

	// Should not get any more outputs
	select {
	case <-debouncer.Output():
		t.Fatal("received unexpected second output")
	case <-time.After(200 * time.Millisecond):
		// Expected: no more outputs
	}
}

func TestDebouncer_ZeroDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	debouncer := New(0)
	debouncer.Start(ctx)

	debouncer.Trigger()

	// Should fire immediately
	select {
	case <-debouncer.Output():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("debouncer did not fire immediately")
	}
}
