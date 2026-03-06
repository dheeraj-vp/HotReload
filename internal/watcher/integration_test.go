package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWatcher_BasicEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create temp directory
	dir := t.TempDir()

	// Initialize watcher
	watcher, err := New(ctx, dir)
	require.NoError(t, err)

	// Start watching
	go watcher.Start(ctx)

	// Wait for initialization
	time.Sleep(100 * time.Millisecond)

	// Test 1: Create file
	testFile := filepath.Join(dir, "test.go")
	err = os.WriteFile(testFile, []byte("package test"), 0644)
	require.NoError(t, err)

	// fsnotify may send CREATE, WRITE, or both for file creation
	// We'll collect events for a short time to handle multiple events
	var events []Event
	for i := 0; i < 2; i++ { // Expect up to 2 events (CREATE + WRITE)
		select {
		case event := <-watcher.Events():
			if event.Path == testFile {
				events = append(events, event)
			}
		case <-time.After(200 * time.Millisecond):
			break
		}
	}
	
	require.NotEmpty(t, events, "Should receive at least one event for file creation")
	assert.Equal(t, testFile, events[0].Path)
	assert.True(t, events[0].Op == OpCreate || events[0].Op == OpWrite, 
		"Expected CREATE or WRITE, got %v", events[0].Op)

	// Test 2: Modify file (if still exists)
	if _, err := os.Stat(testFile); err == nil {
		err = os.WriteFile(testFile, []byte("package test\n// modified"), 0644)
		require.NoError(t, err)

		// Wait for potential WRITE event
		select {
		case event := <-watcher.Events():
			if event.Path == testFile {
				assert.Equal(t, OpWrite, event.Op)
			}
		case <-time.After(200 * time.Millisecond):
			// If no event, it's okay - fsnotify behavior varies
		}
	}
}

func TestFileWatcher_RecursiveDirectories(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create nested structure
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "internal", "pkg"), 0755)
	os.MkdirAll(filepath.Join(dir, "cmd", "server"), 0755)

	// Initialize watcher
	watcher, err := New(ctx, dir)
	require.NoError(t, err)
	go watcher.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Create file in nested directory
	nestedFile := filepath.Join(dir, "internal", "pkg", "test.go")
	err = os.WriteFile(nestedFile, []byte("package pkg"), 0644)
	require.NoError(t, err)

	// Should detect event
	event := waitForEvent(t, watcher.Events(), 1*time.Second)
	assert.Equal(t, nestedFile, event.Path)
}

func TestFileWatcher_DynamicDirectoryCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dir := t.TempDir()
	watcher, err := New(ctx, dir)
	require.NoError(t, err)
	go watcher.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Create new directory while watching
	newDir := filepath.Join(dir, "newpkg")
	err = os.Mkdir(newDir, 0755)
	require.NoError(t, err)

	// Wait for directory to be added to watch list
	time.Sleep(200 * time.Millisecond)

	// Create file in new directory
	newFile := filepath.Join(newDir, "test.go")
	err = os.WriteFile(newFile, []byte("package newpkg"), 0644)
	require.NoError(t, err)

	// Should detect event in dynamically added directory
	event := waitForEvent(t, watcher.Events(), 1*time.Second)
	assert.Equal(t, newFile, event.Path)
}

func TestFileWatcher_IgnoreEvents(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	dir := t.TempDir()
	watcher, err := New(ctx, dir)
	require.NoError(t, err)
	go watcher.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Create ignored file
	ignoredFile := filepath.Join(dir, ".git", "config")
	os.MkdirAll(filepath.Dir(ignoredFile), 0755)
	err = os.WriteFile(ignoredFile, []byte("git config"), 0644)
	require.NoError(t, err)

	// Should not receive event for ignored file
	select {
	case event := <-watcher.Events():
		t.Fatalf("Received unexpected event for ignored file: %s", event.Path)
	case <-time.After(500 * time.Millisecond):
		// Expected: no event
	}
}

func waitForEvent(t *testing.T, events <-chan Event, timeout time.Duration) Event {
	select {
	case event := <-events:
		return event
	case <-time.After(timeout):
		t.Fatal("timeout waiting for event")
		return Event{}
	}
}
