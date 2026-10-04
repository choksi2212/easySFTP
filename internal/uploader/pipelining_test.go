package uploader

import (
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/eiserv/easySFTP/internal/config"

	"github.com/pkg/sftp"
)

// pipeliningProbe wraps the server's FilePut so tests can observe how the
// client writes one file: how many WriteAt calls it made, and how many were
// in flight at the same time. The writer a request server hands out is called
// from several goroutines when the client pipelines (pkg/sftp's worker pool
// fans Write requests out), so the probe must itself be safe under exactly
// the concurrency it is counting.
//
// All non-Put methods delegate to the real in-memory handler untouched.
type pipeliningProbe struct {
	inner sftp.Handlers

	mu          sync.Mutex
	maxInFlight int   // largest number of simultaneous WriteAt calls seen
	inFlight    int   // WriteAt calls currently running
	calls       int   // total WriteAt calls seen
	bytes       int64 // total bytes handed to WriteAt
	delay       time.Duration
}

// Filewrite is FileWriter.Filewrite: return a WriterAt that counts overlap.
func (p *pipeliningProbe) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	inner, err := p.inner.FilePut.Filewrite(r)
	if err != nil {
		return nil, err
	}
	return &countingWriterAt{probe: p, inner: inner}, nil
}

func (p *pipeliningProbe) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	return p.inner.FileGet.Fileread(r)
}

func (p *pipeliningProbe) Filecmd(r *sftp.Request) error {
	return p.inner.FileCmd.Filecmd(r)
}

func (p *pipeliningProbe) PosixRename(r *sftp.Request) error {
	return posixRenamePassthrough(p.inner.FileCmd, r)
}

func (p *pipeliningProbe) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	return p.inner.FileList.Filelist(r)
}

// countingWriterAt counts concurrent WriteAt calls per open remote file and
// delays each one so overlap is visible even on an in-process pipe.
type countingWriterAt struct {
	probe *pipeliningProbe
	inner io.WriterAt
}

func (w *countingWriterAt) WriteAt(p []byte, off int64) (int, error) {
	pr := w.probe
	pr.mu.Lock()
	pr.calls++
	pr.inFlight++
	if pr.inFlight > pr.maxInFlight {
		pr.maxInFlight = pr.inFlight
	}
	pr.mu.Unlock()

	if pr.delay > 0 {
		time.Sleep(pr.delay)
	}
	n, err := w.inner.WriteAt(p, off)

	pr.mu.Lock()
	pr.inFlight--
	pr.bytes += int64(n)
	pr.mu.Unlock()
	return n, err
}

// withPipeliningProbe gives the test server a FilePut wrapper that records
// how the client wrote the file, optionally delaying every WriteAt so
// overlapping requests are observable.
func withPipeliningProbe(pr *pipeliningProbe, delay time.Duration) serverOption {
	return func(s *testServer) {
		pr.inner = s.handlers
		pr.delay = delay
		s.handlers.FilePut = pr
	}
}

// TestUploadPipelinesWritesPerFile is the test issue #276 asks for: for a
// multi-packet file and request_concurrency above 1, more than one write must
// be in flight. Before the fix, uploadFile handed pkg/sftp a reader with no
// Len/Size/Stat, so ReadFrom fell back to its sequential loop and the setting
// moved nothing: every file went out one 32 KiB packet per round-trip.
func TestUploadPipelinesWritesPerFile(t *testing.T) {
	srv := startTestServer(t)

	// Wrap the already-started server's handler set. startTestServer builds
	// InMemHandler() and copies it into srv.handlers before serving; the
	// request server reads that Handlers struct when a session channel opens,
	// so swapping FilePut before the first upload reaches the probe.
	pr := &pipeliningProbe{}
	withPipeliningProbe(pr, 3*time.Millisecond)(srv)

	local := t.TempDir()
	// 2 MiB: 64 packets of 32 KiB, enough to fill a depth-16 pipeline many
	// times over, small enough to keep the memFile write delays fast.
	if err := os.WriteFile(local+"/big.bin", make([]byte, 2<<20), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := baseConfig(srv)
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}
	cfg.Concurrency = 1
	cfg.SftpRequestConcurrency = 64

	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	pr.mu.Lock()
	maxInFlightht, calls, bytes := pr.maxInFlight, pr.calls, pr.bytes
	pr.mu.Unlock()
	// The reader's one-packet lead means even depth 1 can show two
	// simultaneous server-side writes (see the depth-one test below), so
	// "more than one" is not enough to prove the setting reached the wire:
	// anything past the lead is real pipelining. The in-process request
	// server runs 8 workers, so a 64-deep pipeline against it tops out at
	// 8 concurrent WriteAt calls.
	if maxInFlightht < 4 {
		t.Fatalf("upload never pipelined: %d write(s) in flight at most (calls=%d, bytes=%d); request_concurrency=64 must put more than the one-packet lead in flight for a multi-packet file", maxInFlightht, calls, bytes)
	}
	if calls < 64 {
		t.Fatalf("expected at least 64 write requests for a 2 MiB file, got %d", calls)
	}
	if bytes != int64(2<<20) {
		t.Fatalf("wrote %d bytes in total, expected the whole 2 MiB file", bytes)
	}
}

// TestUploadBoundedLeadWhenDepthOne pins the other side: request_concurrency
// of 1 means a single worker waiting for each acknowledgement, so the only
// overlap the server can ever see is pkg/sftp's reader lead: the next packet
// is dispatched to the wire before the previous one's worker slot frees up.
// The bound is the depth plus that one lead: two. A run that shows three at
// depth one would mean the depth is not reaching the write path at all.
func TestUploadBoundedLeadWhenDepthOne(t *testing.T) {
	srv := startTestServer(t)
	pr := &pipeliningProbe{}
	withPipeliningProbe(pr, 3*time.Millisecond)(srv)

	local := t.TempDir()
	if err := os.WriteFile(local+"/big.bin", make([]byte, 512*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := baseConfig(srv)
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}
	cfg.Concurrency = 1
	cfg.SftpRequestConcurrency = 1

	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	pr.mu.Lock()
	maxInFlightht, calls := pr.maxInFlight, pr.calls
	pr.mu.Unlock()
	if maxInFlightht > 2 {
		t.Fatalf("depth-1 upload had %d writes in flight at once; the single worker plus the reader lead bounds it at 2", maxInFlightht)
	}
	if calls < 16 {
		t.Fatalf("expected at least 16 write requests for a 512 KiB file, got %d", calls)
	}
}

// TestStallWatchdogStillFiresUnderPipelining re-runs the stall scenario at a
// real pipeline depth: the server stops reading after 256 KiB while the
// client runs request_concurrency 16. The reader is then gated on
// acknowledgements (pkg/sftp hands packets to workers over an unbuffered
// channel), so reads stop one window after the acks do and the watchdog must
// still fire inside the test's window.
func TestStallWatchdogStillFiresUnderPipelining(t *testing.T) {
	srv := startTestServer(t, withStallAfter(256*1024))

	local := t.TempDir()
	writeTree(t, local, map[string]string{
		"big.bin": string(make([]byte, 4<<20)), // 4 MiB, far beyond the stall point
	})

	cfg := baseConfig(srv)
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}
	cfg.Concurrency = 1
	// The depth the default policy would pick for a large file: the scenario
	// the one-read-lag tick has to stay meaningful in.
	cfg.SftpRequestConcurrency = 16
	cfg.StallTimeout = 1 * time.Second

	start := time.Now()
	_, err := Run(context.Background(), cfg, testLogger{t})
	elapsed := time.Since(start)

	if err == nil || !strings.Contains(err.Error(), "stalled") {
		t.Fatalf("expected a transfer-stalled error, got %v", err)
	}
	if elapsed > 30*time.Second {
		t.Fatalf("run took %s; the stall-timeout should have failed it fast", elapsed)
	}
}
