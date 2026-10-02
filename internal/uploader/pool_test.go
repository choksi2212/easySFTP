package uploader

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/eiserv/easySFTP/internal/autotune"
	"github.com/eiserv/easySFTP/internal/config"
)

// poolTree writes n small files, enough of them that every pool slot is used
// (files are spread over the pool by their plan index).
func poolTree(t *testing.T, root string, n int) {
	t.Helper()
	files := make(map[string]string, n)
	for i := 0; i < n; i++ {
		files[fmt.Sprintf("file%02d.txt", i)] = fmt.Sprintf("content %d", i)
	}
	writeTree(t, root, files)
}

// TestConnectionPoolOpensConfiguredConnections pins the point of issue #158:
// advanced.connections opens that many SSH connections, and the default of one
// keeps the single-connection behavior. The count is read before any
// verifyClient runs, since those dial the same server.
func TestConnectionPoolOpensConfiguredConnections(t *testing.T) {
	for _, tc := range []struct {
		name        string
		connections int
		want        int32
	}{
		{"default is one connection", 0, 1},
		{"explicit single connection", 1, 1},
		{"pool of three", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := startTestServer(t)
			local := t.TempDir()
			poolTree(t, local, 9)

			cfg := baseConfig(srv)
			cfg.Concurrency = 3
			cfg.Connections = tc.connections
			cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}

			stats, err := Run(context.Background(), cfg, testLogger{t})
			if err != nil {
				t.Fatal(err)
			}
			if got := atomic.LoadInt32(&srv.accepted); got != tc.want {
				t.Errorf("expected %d connection(s), got %d", tc.want, got)
			}
			if stats.FilesUploaded != 9 {
				t.Fatalf("expected 9 uploads, got %d", stats.FilesUploaded)
			}
			if got := readRemote(t, srv, "/www/file08.txt"); got != "content 8" {
				t.Errorf("unexpected content: %q", got)
			}
		})
	}
}

// TestConnectionPoolNeverExceedsConcurrency: a connection no worker ever picks
// is a handshake for nothing, so the pool is capped at the number of parallel
// uploads and says so.
func TestConnectionPoolNeverExceedsConcurrency(t *testing.T) {
	srv := startTestServer(t)
	local := t.TempDir()
	poolTree(t, local, 6)

	cfg := baseConfig(srv)
	cfg.Concurrency = 2
	cfg.Connections = 8
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}

	log := &recordingLogger{testLogger: testLogger{t}}
	if _, err := Run(context.Background(), cfg, log); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&srv.accepted); got != 2 {
		t.Errorf("expected the pool to be capped at concurrency (2), got %d connection(s)", got)
	}
	if !containsSubstring(log.infos, "only 2 file(s) upload in parallel") {
		t.Errorf("expected an info line about the cap, got %v", log.infos)
	}
}

// TestConnectionPoolDegradesWhenServerRefusesConnections: sshd's MaxStartups
// and per-account limits on shared hosting are real, so a refused extra
// connection must cost a warning and not the deploy.
func TestConnectionPoolDegradesWhenServerRefusesConnections(t *testing.T) {
	srv := startTestServer(t, withMaxConns(1))
	local := t.TempDir()
	poolTree(t, local, 6)

	cfg := baseConfig(srv)
	cfg.Concurrency = 3
	cfg.Connections = 3
	cfg.Uploads = []config.UploadPair{{Local: local, Remote: "/www"}}

	log := &recordingLogger{testLogger: testLogger{t}}
	stats, err := Run(context.Background(), cfg, log)
	if err != nil {
		t.Fatalf("a server refusing extra connections must not fail the run: %v", err)
	}
	if stats.FilesUploaded != 6 {
		t.Errorf("expected 6 uploads, got %d", stats.FilesUploaded)
	}
	if !containsSubstring(log.warnings, "continues on its first connection") {
		t.Errorf("expected a warning about the refused connection, got %v", log.warnings)
	}
}

// TestGrantedCountsTheConnectionsTheServerGave pins the other half of a
// refusal: what the run remembers about the server must be how many
// connections it was actually granted, not how many pool slots the run
// touched. A refused slot is aliased to the first connection (see acquire), so
// a walk that counts non-nil slots answers "the spread" and the next run asks
// for the same refused connections again (issue #281).
//
// The workers start together, the way a real upload phase does, so several
// slots are already waiting for their own connection when the refusal lands
// and they are aliased as well: withMaxConns(2) grants the initial connection
// plus one dial, and everything past that must count as one, not as its slot.
func TestGrantedCountsTheConnectionsTheServerGave(t *testing.T) {
	srv := startTestServer(t, withMaxConns(2))
	cfg := autoConfig(srv)

	tune := newTuning(cfg)
	tune.resolveRunWide(autotune.Workload{Uploads: 100, UploadBytes: 1 << 20, LargestUpload: 1 << 12})

	log := &recordingLogger{testLogger: testLogger{t}}
	sess, err := newSession(context.Background(), cfg, tune, log)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.close()

	if got := sess.setSpread(4); got != 4 {
		t.Fatalf("setSpread(4) = %d, want the pool to widen before anything is dialed", got)
	}
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if client, _, _ := sess.acquire(i); client == nil {
				t.Errorf("acquire(%d) returned no client", i)
			}
		}()
	}
	wg.Wait()

	granted, refused := sess.granted()
	if !refused {
		t.Fatal("the server refused a dial, but the session does not remember it")
	}
	if granted != 2 {
		t.Errorf("granted() = %d, want 2: the initial connection plus the one extra the server allowed", granted)
	}
}

// TestGrantedCountsAliasesWithPinnedConnections is the same walk with
// advanced.connections pinned: the refusal is phrased differently and the
// spread is not pulled back (a pinned value is never clamped), but the number
// the run remembers still has to be the grant, because the cache writes it
// either way.
func TestGrantedCountsAliasesWithPinnedConnections(t *testing.T) {
	srv := startTestServer(t, withMaxConns(1))
	cfg := baseConfig(srv)
	cfg.Connections = 4

	tune := newTuning(cfg)
	tune.resolveRunWide(autotune.Workload{Uploads: 100, UploadBytes: 1 << 20, LargestUpload: 1 << 12})

	log := &recordingLogger{testLogger: testLogger{t}}
	sess, err := newSession(context.Background(), cfg, tune, log)
	if err != nil {
		t.Fatal(err)
	}
	defer sess.close()

	if got := sess.setSpread(4); got != 4 {
		t.Fatalf("setSpread(4) = %d, want the pool to widen before anything is dialed", got)
	}
	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if client, _, _ := sess.acquire(i); client == nil {
				t.Errorf("acquire(%d) returned no client", i)
			}
		}()
	}
	wg.Wait()

	granted, refused := sess.granted()
	if !refused {
		t.Fatal("the server refused a dial, but the session does not remember it")
	}
	if granted != 1 {
		t.Errorf("granted() = %d, want 1: only the initial connection was allowed", granted)
	}
}

func containsSubstring(lines []string, want string) bool {
	for _, l := range lines {
		if strings.Contains(l, want) {
			return true
		}
	}
	return false
}
