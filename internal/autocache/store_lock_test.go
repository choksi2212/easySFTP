package autocache

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The Windows runner saw the delete-pending denial on 2026-10-04 in PR #301:
// TestConcurrentUpdateFileMergesEveryWriter failed with "open
// ...auto.json.lock: Access is denied" while the six writers handed the
// sidecar over. These tests fault lockCreate the same way the platform does,
// deterministically, on every OS.

func withLockTimings(t *testing.T, wait, poll time.Duration) {
	t.Helper()
	origWait, origPoll := lockWait, lockPoll
	lockWait, lockPoll = wait, poll
	t.Cleanup(func() { lockWait, lockPoll = origWait, origPoll })
}

func withFaultedLockCreate(t *testing.T, fault func(lockPath string, underlying func(string) (*os.File, error)) (*os.File, error)) {
	t.Helper()
	underlying := lockCreate
	lockCreate = func(lockPath string) (*os.File, error) {
		return fault(lockPath, underlying)
	}
	t.Cleanup(func() { lockCreate = underlying })
}

// TestDeniedLockCreateIsPolledNotFailed: the first create is denied the way a
// delete-pending sidecar denies it on Windows, and the second attempt gets
// the lock. The write must go through, proving the denial was polled rather
// than returned. Away from Windows the same denial is returned immediately,
// which the fail-fast test below covers.
func TestDeniedLockCreateIsPolledNotFailed(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows treats a denied create as busy; elsewhere it fails fast")
	}
	path := filepath.Join(t.TempDir(), "auto.json")
	withLockTimings(t, time.Second, time.Millisecond)
	calls := 0
	withFaultedLockCreate(t, func(lockPath string, underlying func(string) (*os.File, error)) (*os.File, error) {
		calls++
		if calls == 1 {
			return nil, os.ErrPermission
		}
		return underlying(lockPath)
	})
	if err := Save(path, stored(mediumTree(), measuredLink())); err != nil {
		t.Fatalf("a denied first create must be polled, not failed: %v", err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("the store did not survive the denied create: %v", err)
	}
}

// TestPersistentDeniedLockCreateWaitsForTheDeadline: on Windows a denial that
// never clears must surface at the deadline, and the error must name the
// sidecar and carry the create error, so a genuinely unwritable directory is
// diagnosable from the cache warning alone. Away from Windows the same
// denial is returned immediately by the fail-fast test below.
func TestPersistentDeniedLockCreateWaitsForTheDeadline(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only Windows treats a denied create as busy; elsewhere it fails fast")
	}
	path := filepath.Join(t.TempDir(), "auto.json")
	withLockTimings(t, 50*time.Millisecond, 5*time.Millisecond)
	withFaultedLockCreate(t, func(lockPath string, underlying func(string) (*os.File, error)) (*os.File, error) {
		return nil, os.ErrPermission
	})
	start := time.Now()
	err := Save(path, stored(mediumTree(), measuredLink()))
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("a denial that persists past lockWait must be an error")
	}
	if elapsed < 40*time.Millisecond {
		t.Errorf("the denial was failed fast after %v, want it waited out lockWait", elapsed)
	}
	if !strings.Contains(err.Error(), "timed out waiting for concurrent cache writer") {
		t.Errorf("the deadline error must say what was being waited for, got: %v", err)
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("the deadline error must carry the create error for diagnosis, got: %v", err)
	}
}

// TestDeniedLockCreateIsFailFastAwayFromWindows: everywhere but Windows a
// denied create is a real permission problem, not a busy signal, and must
// come back immediately rather than burning lockWait polling it.
func TestDeniedLockCreateIsFailFastAwayFromWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows polls the denial as busy; see the two tests above")
	}
	path := filepath.Join(t.TempDir(), "auto.json")
	withLockTimings(t, time.Second, time.Millisecond)
	withFaultedLockCreate(t, func(lockPath string, underlying func(string) (*os.File, error)) (*os.File, error) {
		return nil, os.ErrPermission
	})
	start := time.Now()
	err := Save(path, stored(mediumTree(), measuredLink()))
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("a denied create away from Windows must fail, not succeed")
	}
	if !errors.Is(err, os.ErrPermission) {
		t.Errorf("the create error must be returned as is, got: %v", err)
	}
	if strings.Contains(err.Error(), "timed out waiting for concurrent cache writer") {
		t.Errorf("a real permission problem must not be reported as a busy lock: %v", err)
	}
	if elapsed >= time.Second {
		t.Errorf("the denial was polled for %v, want fail fast", elapsed)
	}
}
