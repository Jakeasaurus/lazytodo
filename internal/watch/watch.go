package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ChangeReason indicates why a file change event was triggered
type ChangeReason string

const (
	ReasonWrite   ChangeReason = "write"
	ReasonChmod   ChangeReason = "chmod"
	ReasonReplace ChangeReason = "replace"
	ReasonCreate  ChangeReason = "create"
)

// Config holds configuration for the file watcher
type Config struct {
	Path         string        // Path to watch (e.g., ~/todo.txt)
	Debounce     time.Duration // Time to wait before firing events (default: 300ms)
	SelfWriteTTL time.Duration // Time to suppress events after self-writes (default: 500ms)
	PollFallback time.Duration // Polling interval if fsnotify fails (0 disables, default: 2s)
	Logger       func(msg string, args ...interface{}) // Optional logger
}

// Runner watches a file and calls onChange when it's modified externally
type Runner interface {
	Start(ctx context.Context, onChange func(ChangeReason)) error
	Stop() error
	BeginSelfWrite() func() // Returns done() function to end suppression
}

type runner struct {
	cfg           Config
	w             *fsnotify.Watcher
	filePath      string
	dirPath       string
	fileName      string
	debounceTimer *time.Timer
	mu            sync.Mutex
	closed        chan struct{}
	selfUntil     atomic.Int64 // Unix nano deadline for self-write suppression
	lastStat      fileInfo     // For polling fallback
}

type fileInfo struct {
	size  int64
	mtime time.Time
}

// New creates a new file watcher with the given configuration
func New(cfg Config) (Runner, error) {
	// Set defaults
	if cfg.Debounce == 0 {
		cfg.Debounce = 300 * time.Millisecond
	}
	if cfg.SelfWriteTTL == 0 {
		cfg.SelfWriteTTL = 500 * time.Millisecond
	}
	if cfg.PollFallback == 0 {
		cfg.PollFallback = 2 * time.Second
	}

	// Expand path (handle ~ and environment variables)
	path, err := expandPath(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path %q: %w", cfg.Path, err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}

	dirPath := filepath.Dir(absPath)
	fileName := filepath.Base(absPath)

	r := &runner{
		cfg:      cfg,
		filePath: absPath,
		dirPath:  dirPath,
		fileName: fileName,
		closed:   make(chan struct{}),
	}

	// Try to create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		r.logf("fsnotify unavailable, will use polling: %v", err)
		// Will use polling fallback
	} else {
		r.w = watcher
	}

	// Initialize file stat for polling
	if stat, err := os.Stat(r.filePath); err == nil {
		r.lastStat = fileInfo{
			size:  stat.Size(),
			mtime: stat.ModTime(),
		}
	}

	return r, nil
}

// Start begins watching the file for changes
func (r *runner) Start(ctx context.Context, onChange func(ChangeReason)) error {
	if r.w != nil {
		return r.startFsnotify(ctx, onChange)
	}
	return r.startPolling(ctx, onChange)
}

func (r *runner) startFsnotify(ctx context.Context, onChange func(ChangeReason)) error {
	// Add directory to watcher (for atomic replacements)
	if err := r.w.Add(r.dirPath); err != nil {
		r.logf("failed to watch directory %s, falling back to polling: %v", r.dirPath, err)
		return r.startPolling(ctx, onChange)
	}

	// Add file to watcher if it exists
	if _, err := os.Stat(r.filePath); err == nil {
		if err := r.w.Add(r.filePath); err != nil {
			r.logf("failed to watch file %s: %v", r.filePath, err)
		}
	}

	r.logf("watching %s with fsnotify", r.filePath)

	go r.fsnotifyLoop(ctx, onChange)
	return nil
}

func (r *runner) fsnotifyLoop(ctx context.Context, onChange func(ChangeReason)) {
	defer r.w.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.closed:
			return
		case event, ok := <-r.w.Events:
			if !ok {
				return
			}
			r.handleEvent(event, onChange)
		case err, ok := <-r.w.Errors:
			if !ok {
				return
			}
			r.logf("watcher error: %v", err)
		}
	}
}

func (r *runner) handleEvent(event fsnotify.Event, onChange func(ChangeReason)) {
	// Check if this is our target file or directory
	isTargetFile := event.Name == r.filePath
	isTargetInDir := filepath.Dir(event.Name) == r.dirPath && filepath.Base(event.Name) == r.fileName

	if !isTargetFile && !isTargetInDir {
		return
	}

	// Check self-write suppression
	if r.isSupressed() {
		r.logf("suppressing event %v (self-write)", event)
		return
	}

	var reason ChangeReason
	if event.Has(fsnotify.Write) {
		reason = ReasonWrite
	} else if event.Has(fsnotify.Chmod) {
		reason = ReasonChmod
	} else if event.Has(fsnotify.Create) {
		reason = ReasonCreate
		// If the file was created, add it to the watcher
		if isTargetInDir {
			r.w.Add(r.filePath)
		}
	} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		reason = ReasonReplace
		// File was removed/renamed, remove from watcher
		if isTargetFile {
			r.w.Remove(r.filePath)
		}
	} else {
		return
	}

	r.logf("file event: %v -> %s", event, reason)
	r.debounceOnChange(reason, onChange)
}

func (r *runner) startPolling(ctx context.Context, onChange func(ChangeReason)) error {
	r.logf("watching %s with polling (interval: %v)", r.filePath, r.cfg.PollFallback)

	go r.pollingLoop(ctx, onChange)
	return nil
}

func (r *runner) pollingLoop(ctx context.Context, onChange func(ChangeReason)) {
	ticker := time.NewTicker(r.cfg.PollFallback)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-r.closed:
			return
		case <-ticker.C:
			r.checkPolling(onChange)
		}
	}
}

func (r *runner) checkPolling(onChange func(ChangeReason)) {
	if r.isSupressed() {
		return
	}

	stat, err := os.Stat(r.filePath)
	if err != nil {
		// File doesn't exist or can't be accessed
		if !r.lastStat.mtime.IsZero() {
			// File was removed
			r.lastStat = fileInfo{}
			r.debounceOnChange(ReasonReplace, onChange)
		}
		return
	}

	newStat := fileInfo{
		size:  stat.Size(),
		mtime: stat.ModTime(),
	}

	// Check if file changed
	if r.lastStat.mtime.IsZero() {
		// File was created
		r.lastStat = newStat
		r.debounceOnChange(ReasonCreate, onChange)
	} else if newStat.mtime.After(r.lastStat.mtime) || newStat.size != r.lastStat.size {
		// File was modified
		r.lastStat = newStat
		r.debounceOnChange(ReasonWrite, onChange)
	}
}

func (r *runner) debounceOnChange(reason ChangeReason, onChange func(ChangeReason)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Reset or create debounce timer
	if r.debounceTimer != nil {
		r.debounceTimer.Stop()
	}

	r.debounceTimer = time.AfterFunc(r.cfg.Debounce, func() {
		r.logf("firing onChange: %s", reason)
		onChange(reason)
	})
}

// BeginSelfWrite marks the start of a self-write operation
func (r *runner) BeginSelfWrite() func() {
	deadline := time.Now().Add(r.cfg.SelfWriteTTL).UnixNano()
	r.selfUntil.Store(deadline)

	return func() {
		// End suppression early
		r.selfUntil.Store(time.Now().UnixNano())
	}
}

func (r *runner) isSupressed() bool {
	return time.Now().UnixNano() < r.selfUntil.Load()
}

// Stop stops the file watcher
func (r *runner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	close(r.closed)

	if r.debounceTimer != nil {
		r.debounceTimer.Stop()
	}

	if r.w != nil {
		return r.w.Close()
	}

	return nil
}

func (r *runner) logf(format string, args ...interface{}) {
	if r.cfg.Logger != nil {
		r.cfg.Logger(format, args...)
	}
}

// expandPath expands ~ and environment variables in a file path
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return os.ExpandEnv(path), nil
}