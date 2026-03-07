package crashguard

import (
	"sync"
	"time"
)

type CrashGuard struct {
	mu           sync.Mutex
	crashCount   int
	lastCrash    time.Time
	backoffDelay time.Duration
	maxRestarts  int
	baseDelay    time.Duration
	maxDelay     time.Duration
	window       time.Duration
}

type Config struct {
	MaxRestarts  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Window      time.Duration
}

func New(config Config) *CrashGuard {
	if config.MaxRestarts == 0 {
		config.MaxRestarts = 10
	}
	if config.BaseDelay == 0 {
		config.BaseDelay = 1 * time.Second
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.Window == 0 {
		config.Window = 5 * time.Minute
	}

	return &CrashGuard{
		maxRestarts:  config.MaxRestarts,
		backoffDelay: config.BaseDelay,
		baseDelay:    config.BaseDelay,
		maxDelay:     config.MaxDelay,
		window:       config.Window,
	}
}

func (cg *CrashGuard) ShouldRestart() bool {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	now := time.Now()

	if !cg.lastCrash.IsZero() && now.Sub(cg.lastCrash) > cg.window {
		cg.crashCount = 0
	}

	if cg.crashCount >= cg.maxRestarts {
		return false
	}

	return true
}

func (cg *CrashGuard) RecordCrash() {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	cg.crashCount++
	cg.lastCrash = time.Now()

	if cg.crashCount == 1 {
		cg.backoffDelay = cg.baseDelay
	} else if cg.crashCount == 2 {
		cg.backoffDelay = 2 * cg.baseDelay
	} else if cg.crashCount == 3 {
		cg.backoffDelay = 4 * cg.baseDelay
	} else if cg.crashCount == 4 {
		cg.backoffDelay = 8 * cg.baseDelay
	} else if cg.crashCount == 5 {
		cg.backoffDelay = 16 * cg.baseDelay
	} else {
		cg.backoffDelay = 32 * cg.baseDelay
		if cg.backoffDelay > cg.maxDelay {
			cg.backoffDelay = cg.maxDelay
		}
	}
}

func (cg *CrashGuard) GetBackoffDelay() time.Duration {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	return cg.backoffDelay
}

func (cg *CrashGuard) Reset() {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	cg.crashCount = 0
	cg.lastCrash = time.Time{}
	cg.backoffDelay = cg.baseDelay
}

func (cg *CrashGuard) GetStats() Stats {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	return Stats{
		CrashCount:   cg.crashCount,
		LastCrash:    cg.lastCrash,
		BackoffDelay: cg.backoffDelay,
		CanRestart:   cg.crashCount < cg.maxRestarts,
	}
}

type Stats struct {
	CrashCount   int
	LastCrash    time.Time
	BackoffDelay time.Duration
	CanRestart   bool
}
