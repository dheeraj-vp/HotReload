package watcher

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	fsWatcher *fsnotify.Watcher
	root      string
	events    chan Event
	errors    chan error
}

type Event struct {
	Path      string
	Op        Operation
	Timestamp time.Time
}

type Operation int

const (
	OpCreate Operation = iota
	OpWrite
	OpRemove
	OpRename
)

func New(ctx context.Context, root string) (*FileWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fw := &FileWatcher{
		fsWatcher: fsWatcher,
		root:      root,
		events:    make(chan Event, 100),
		errors:    make(chan error, 10),
	}

	if err := fw.addRecursive(root); err != nil {
		fsWatcher.Close()
		return nil, err
	}

	slog.Info("watcher initialized", "paths_watched", fw.getWatchCount())
	return fw, nil
}

func (fw *FileWatcher) Start(ctx context.Context) error {
	go func() {
		defer close(fw.events)
		defer close(fw.errors)
		defer fw.fsWatcher.Close()

		for {
			select {
			case <-ctx.Done():
				slog.Debug("watcher stopping", "reason", "context cancelled")
				return

			case fsEvent, ok := <-fw.fsWatcher.Events:
				if !ok {
					return
				}
				fw.handleFsnotifyEvent(fsEvent)

			case err, ok := <-fw.fsWatcher.Errors:
				if !ok {
					return
				}
				select {
				case fw.errors <- err:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func (fw *FileWatcher) Events() <-chan Event {
	return fw.events
}

func (fw *FileWatcher) Errors() <-chan error {
	return fw.errors
}

func (fw *FileWatcher) addRecursive(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ShouldIgnore(walkPath) {
			if info.IsDir() {
				slog.Debug("skipping directory", "path", walkPath)
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			if err := fw.fsWatcher.Add(walkPath); err != nil {
				slog.Warn("failed to watch directory", "path", walkPath, "error", err)
				return nil
			}
			slog.Debug("watching directory", "path", walkPath)
		}

		return nil
	})
}

func (fw *FileWatcher) handleFsnotifyEvent(fsEvent fsnotify.Event) {
	if ShouldIgnore(fsEvent.Name) {
		slog.Debug("ignoring event", "path", fsEvent.Name, "op", fsEvent.Op)
		return
	}

	var op Operation
	switch {
	case fsEvent.Op&fsnotify.Create == fsnotify.Create:
		op = OpCreate
		if info, err := os.Stat(fsEvent.Name); err == nil && info.IsDir() {
			fw.handleDirCreate(fsEvent.Name)
			return
		}
	case fsEvent.Op&fsnotify.Write == fsnotify.Write:
		op = OpWrite
	case fsEvent.Op&fsnotify.Remove == fsnotify.Remove:
		op = OpRemove
	case fsEvent.Op&fsnotify.Rename == fsnotify.Rename:
		op = OpRename
	default:
		return
	}

	if info, err := os.Stat(fsEvent.Name); err == nil && !info.IsDir() {
		event := Event{
			Path:      fsEvent.Name,
			Op:        op,
			Timestamp: time.Now(),
		}

		select {
		case fw.events <- event:
			slog.Debug("file event", "path", event.Path, "op", op)
		default:
			slog.Warn("event channel full, dropping event", "path", fsEvent.Name)
		}
	}
}

func (fw *FileWatcher) handleDirCreate(path string) error {
	if err := fw.addRecursive(path); err != nil {
		slog.Warn("failed to watch new directory", "path", path, "error", err)
		return err
	}
	slog.Debug("added new directory to watch list", "path", path)
	return nil
}

func (fw *FileWatcher) getWatchCount() int {
	count := 0
	filepath.Walk(fw.root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() && !ShouldIgnore(path) {
			count++
		}
		return nil
	})
	return count
}
