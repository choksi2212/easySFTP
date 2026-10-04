package autocache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// fileMode is what the cache file is created with. It holds no secrets (the
// target is a fingerprint, see Record.Target), but it describes somebody's
// deployment and there is no reason for it to be world readable.
const fileMode = 0o600

// The lock timings are vars so a test can shrink the wait instead of
// sleeping through the real three seconds.
var (
	lockWait  = 3 * time.Second
	lockStale = 10 * time.Minute
	lockPoll  = 25 * time.Millisecond
)

// ErrEmpty reports a path that holds no usable store yet: the file does not
// exist, or it is empty. It is the normal state of a first run and the caller
// says so at debug level rather than warning about it.
var ErrEmpty = errors.New("no auto-tuning cache yet")

// Load reads the store at path.
//
// Every failure mode here is recoverable by design, because the cache is an
// optimisation and must never be able to fail a deploy (issue #212). A missing
// file is ErrEmpty. A file that cannot be parsed, or that was written by a
// newer easySFTP, comes back as an error together with an empty usable store,
// so the caller can warn once and carry on with the built-in prior. The one
// thing Load will not do is guess: a version it does not know is not read
// field by field in the hope that the fields still mean what they meant.
func Load(path string) (*Store, error) {
	if path == "" {
		return &Store{Version: FormatVersion}, ErrEmpty
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Store{Version: FormatVersion}, ErrEmpty
		}
		return &Store{Version: FormatVersion}, err
	}
	if len(data) == 0 {
		return &Store{Version: FormatVersion}, ErrEmpty
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return &Store{Version: FormatVersion}, fmt.Errorf("parsing %s: %w", path, err)
	}
	if s.Version != FormatVersion {
		return &Store{Version: FormatVersion}, fmt.Errorf(
			"%s is version %d and this easySFTP writes version %d", path, s.Version, FormatVersion)
	}
	s.Version = FormatVersion
	return &s, nil
}

// Save writes the store to path, creating the directory if it is missing (a
// cache path pointed at a directory a CI cache has not restored yet is the
// normal first run, not an error).
//
// The write is atomic: a temporary sibling, then a rename. A cache file
// truncated by a cancelled job would be read back as corrupt on the next run,
// which costs a warning and a re-measure; it is cheap to make that impossible.
func Save(path string, s *Store) error {
	if path == "" || s == nil {
		return nil
	}
	return withStoreLock(path, func() error { return saveUnlocked(path, s) })
}

// UpdateFile serializes one run's read-modify-write with every other process
// using the same cache path. Reloading after the lock is acquired is the
// important part: each writer merges its observation into the records the
// previous writer just committed instead of replacing the whole document it
// happened to read at startup.
func UpdateFile(path string, obs Observation, at time.Time) (bool, error) {
	if path == "" {
		return false, nil
	}
	changed := false
	err := withStoreLock(path, func() error {
		store, loadErr := Load(path)
		if loadErr != nil && !errors.Is(loadErr, ErrEmpty) {
			// The caller already warned when it read this unusable file. A fresh
			// valid store is the recoverable outcome at write-back time.
			store = &Store{Version: FormatVersion}
		}
		changed = store.Update(obs, at)
		if !changed {
			return nil
		}
		return saveUnlocked(path, store)
	})
	return changed, err
}

func saveUnlocked(path string, s *Store) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	s.Version = FormatVersion
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name) // no-op once the rename succeeded
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, fileMode); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// withStoreLock is a small cross-platform lock built from O_EXCL. The lock is
// held only around a tiny JSON read and atomic write. A killed process can
// leave the sidecar behind, so an old lock is recoverable; a fresh lock that
// does not clear within lockWait costs a cache warning, never the deploy.
//
// On Windows a create can also be denied outright while the lock is being
// removed: the previous holder's delete has marked the sidecar delete-pending,
// the name only disappears once the last share-delete handle closes, and a
// create that lands in that window gets ERROR_ACCESS_DENIED instead of
// ERROR_FILE_EXISTS. A denial there is polled like a busy lock; one that
// persists past lockWait surfaces at the deadline with the create error.
func withStoreLock(path string, fn func() error) error {
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return lockPollLoop(path+".lock", fn)
}

// lockCreate is the lock acquisition primitive, split out so tests can fault
// the create the way the platform does, without creating the sidecar.
var lockCreate = func(lockPath string) (*os.File, error) {
	return os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
}

// lockCreateDeniedIsBusy reports whether a denied lock create means the lock
// is taken and should be polled, instead of a permission problem worth
// failing fast on. Only Windows does this: a create that races the previous
// holder's remove of the sidecar is denied outright, because the remove has
// left the name delete-pending and the name only disappears once the last
// share-delete handle (an antivirus scan, a backup agent) closes. Everywhere
// else a denied create is a real permission problem, so it is returned as is.
func lockCreateDeniedIsBusy(err error) bool {
	return runtime.GOOS == "windows" && errors.Is(err, os.ErrPermission)
}

// lockPollLoop is the poll loop withStoreLock's comment describes, taking the
// resolved sidecar path. ErrExist is the normal busy signal; on Windows a
// denied create (the delete-pending race above) is polled the same way.
func lockPollLoop(lockPath string, fn func() error) error {
	deadline := time.Now().Add(lockWait)
	for {
		lock, err := lockCreate(lockPath)
		if err == nil {
			name := lock.Name()
			if _, writeErr := fmt.Fprintf(lock, "%d\n", os.Getpid()); writeErr != nil {
				lock.Close()
				os.Remove(name)
				return writeErr
			}
			if closeErr := lock.Close(); closeErr != nil {
				os.Remove(name)
				return closeErr
			}
			defer os.Remove(name)
			return fn()
		}
		if !errors.Is(err, os.ErrExist) && !lockCreateDeniedIsBusy(err) {
			return err
		}
		if info, statErr := os.Stat(lockPath); statErr == nil && time.Since(info.ModTime()) > lockStale {
			if removeErr := os.Remove(lockPath); removeErr == nil || errors.Is(removeErr, os.ErrNotExist) {
				continue
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for concurrent cache writer at %s: %w", lockPath, err)
		}
		time.Sleep(lockPoll)
	}
}
