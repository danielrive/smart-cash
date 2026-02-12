package secrets

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// DefaultSecretMountPath is the default path where secrets are mounted by Secrets Store CSI Driver
	DefaultSecretMountPath = "/mnt/secrets-store"
	// DefaultSecretFileName is the default filename for the JWT secret
	DefaultSecretFileName = "secret"
)

var (
	secretCache     = make(map[string][]byte)
	secretCacheLock sync.RWMutex
	watchers        = make(map[string]*secretWatcher)
	watchersLock    sync.Mutex
)

type secretWatcher struct {
	path     string
	callback func([]byte) error
	stop     chan struct{}
	done     chan struct{}
}

// ReadSecret reads a secret from the mounted file path.
// It reads from /mnt/secrets-store/{filename} by default.
func ReadSecret(filename string) ([]byte, error) {
	if filename == "" {
		filename = DefaultSecretFileName
	}

	path := filepath.Join(DefaultSecretMountPath, filename)

	// Check cache first
	secretCacheLock.RLock()
	if cached, ok := secretCache[path]; ok {
		secretCacheLock.RUnlock()
		return cached, nil
	}
	secretCacheLock.RUnlock()

	// Read from file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from %s: %w", path, err)
	}

	// Update cache
	secretCacheLock.Lock()
	secretCache[path] = data
	secretCacheLock.Unlock()

	return data, nil
}

// WatchSecret watches a secret file for changes and calls the callback when it changes.
// This enables zero-downtime secret rotation.
func WatchSecret(ctx context.Context, filename string, callback func([]byte) error) error {
	if filename == "" {
		filename = DefaultSecretFileName
	}

	path := filepath.Join(DefaultSecretMountPath, filename)

	watchersLock.Lock()
	defer watchersLock.Unlock()

	// Check if watcher already exists
	if _, exists := watchers[path]; exists {
		return fmt.Errorf("watcher already exists for %s", path)
	}

	watcher := &secretWatcher{
		path:     path,
		callback: callback,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}

	watchers[path] = watcher

	go watcher.watch(ctx)

	return nil
}

// StopWatching stops watching a secret file.
func StopWatching(filename string) {
	if filename == "" {
		filename = DefaultSecretFileName
	}

	path := filepath.Join(DefaultSecretMountPath, filename)

	watchersLock.Lock()
	defer watchersLock.Unlock()

	if watcher, exists := watchers[path]; exists {
		close(watcher.stop)
		<-watcher.done
		delete(watchers, path)
	}
}

func (w *secretWatcher) watch(ctx context.Context) {
	defer close(w.done)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastModTime := time.Time{}
	lastData := []byte{}

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			// Check file modification time
			var stat unix.Stat_t
			if err := unix.Stat(w.path, &stat); err != nil {
				continue
			}

			modTime := time.Unix(stat.Mtim.Sec, stat.Mtim.Nsec)
			if modTime.Equal(lastModTime) {
				continue
			}

			// Read file
			data, err := os.ReadFile(w.path)
			if err != nil {
				continue
			}

			// Check if content changed
			if string(data) == string(lastData) {
				lastModTime = modTime
				continue
			}

			// Update cache
			secretCacheLock.Lock()
			secretCache[w.path] = data
			secretCacheLock.Unlock()

			// Call callback
			if err := w.callback(data); err != nil {
				// Log error but continue watching
				continue
			}

			lastModTime = modTime
			lastData = data
		}
	}
}

// ReadJWTSecret is a convenience function to read the JWT secret.
func ReadJWTSecret() ([]byte, error) {
	return ReadSecret(DefaultSecretFileName)
}
