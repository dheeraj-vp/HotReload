package crashguard

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCrashGuard_ShouldRestart(t *testing.T) {
	config := Config{
		MaxRestarts: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Window:     5 * time.Minute,
	}
	cg := New(config)

	assert.True(t, cg.ShouldRestart())

	cg.RecordCrash()
	assert.True(t, cg.ShouldRestart())

	cg.RecordCrash()
	assert.True(t, cg.ShouldRestart())

	cg.RecordCrash()
	assert.False(t, cg.ShouldRestart())
}

func TestCrashGuard_ExponentialBackoff(t *testing.T) {
	config := Config{
		MaxRestarts: 5,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
	}
	cg := New(config)

	cg.RecordCrash()
	assert.Equal(t, 1*time.Second, cg.GetBackoffDelay())

	cg.RecordCrash()
	assert.Equal(t, 2*time.Second, cg.GetBackoffDelay())

	cg.RecordCrash()
	assert.Equal(t, 4*time.Second, cg.GetBackoffDelay())

	cg.RecordCrash()
	assert.Equal(t, 8*time.Second, cg.GetBackoffDelay())

	cg.RecordCrash()
	assert.Equal(t, 16*time.Second, cg.GetBackoffDelay())
}

func TestCrashGuard_MaxBackoffLimit(t *testing.T) {
	config := Config{
		MaxRestarts: 10,
		BaseDelay:  1 * time.Second,
		MaxDelay:   10 * time.Second,
	}
	cg := New(config)

	for i := 0; i < 10; i++ {
		cg.RecordCrash()
	}
	assert.Equal(t, 10*time.Second, cg.GetBackoffDelay())
}

func TestCrashGuard_Reset(t *testing.T) {
	config := Config{
		MaxRestarts: 2,
		BaseDelay:  1 * time.Second,
	}
	cg := New(config)

	cg.RecordCrash()
	cg.RecordCrash()
	assert.False(t, cg.ShouldRestart())

	cg.Reset()
	assert.True(t, cg.ShouldRestart())
	assert.Equal(t, 0, cg.crashCount)
	assert.Equal(t, 1*time.Second, cg.GetBackoffDelay())
}

func TestCrashGuard_TimeWindowReset(t *testing.T) {
	config := Config{
		MaxRestarts: 1,
		BaseDelay:  1 * time.Second,
		Window:     100 * time.Millisecond,
	}
	cg := New(config)

	cg.RecordCrash()
	assert.False(t, cg.ShouldRestart())

	time.Sleep(150 * time.Millisecond)

	assert.True(t, cg.ShouldRestart())
	assert.Equal(t, 0, cg.crashCount)
}

func TestCrashGuard_GetStats(t *testing.T) {
	config := Config{
		MaxRestarts: 5,
		BaseDelay:  2 * time.Second,
	}
	cg := New(config)

	stats := cg.GetStats()
	assert.Equal(t, 0, stats.CrashCount)
	assert.True(t, stats.LastCrash.IsZero())
	assert.Equal(t, 2*time.Second, stats.BackoffDelay)
	assert.True(t, stats.CanRestart)

	cg.RecordCrash()
	stats = cg.GetStats()
	assert.Equal(t, 1, stats.CrashCount)
	assert.False(t, stats.LastCrash.IsZero())
	assert.Equal(t, 2*time.Second, stats.BackoffDelay)
	assert.True(t, stats.CanRestart)
}

func TestCrashGuard_DefaultConfig(t *testing.T) {
	cg := New(Config{})

	assert.True(t, cg.ShouldRestart())
	assert.Equal(t, 1*time.Second, cg.GetBackoffDelay())
	
	for i := 0; i < 9; i++ {
		cg.RecordCrash()
		assert.True(t, cg.ShouldRestart())
	}
	
	cg.RecordCrash()
	assert.False(t, cg.ShouldRestart())
}
