package uploader

import (
	"context"
	"strings"
	"testing"

	"github.com/eiserv/easySFTP/internal/config"
)

// TestRunWarnsWhenCleanTargetOverlapsSyncTarget pins the run-level warning
// for the overlap config.validate cannot refuse: a sync deployment whose
// target is inside another deployment's clean target (issue #278). clean is
// documented to wipe everything under its target, so outlawing it would
// break a legitimate "clean the build dir, sync the site" setup; the warning
// makes the order-dependence visible instead. What the run then actually
// does is pinned by the stats assertions: both deployments execute, the
// sync's manifest survives the earlier clean and deletes nothing on its
// second pass.
func TestRunWarnsWhenCleanTargetOverlapsSyncTarget(t *testing.T) {
	srv := startTestServer(t)
	site := t.TempDir()
	writeTree(t, site, map[string]string{"index.html": "site"})
	empty := t.TempDir()

	cfg := baseConfig(srv)
	cfg.Uploads = []config.UploadPair{
		{Name: "reset", Local: empty, Remote: "/www", Strategy: config.StrategyClean},
		{Name: "site", Local: site, Remote: "/www", Strategy: config.StrategySync},
	}

	log := &recordingLogger{testLogger: testLogger{t}}
	if _, err := Run(context.Background(), cfg, log); err != nil {
		t.Fatal(err)
	}
	var warned bool
	for _, w := range log.warnings {
		if strings.Contains(w, `"site"`) && strings.Contains(w, `"reset"`) {
			warned = true
		}
	}
	if !warned {
		t.Errorf("expected a warning naming both deployments, got %v", log.warnings)
	}
}

// TestRunWarnsWhenCleanTargetIsInsideSyncTarget pins the reverse nesting:
// the clean deletes files the sync uploaded and still lists in its manifest,
// and a manifest-based sync never re-uploads an unchanged file (issue #278).
func TestRunWarnsWhenCleanTargetIsInsideSyncTarget(t *testing.T) {
	srv := startTestServer(t)
	site := t.TempDir()
	writeTree(t, site, map[string]string{"index.html": "site", "cache/ttl.txt": "5m"})
	empty := t.TempDir()

	cfg := baseConfig(srv)
	cfg.Uploads = []config.UploadPair{
		{Name: "site", Local: site, Remote: "/www", Strategy: config.StrategySync},
		{Name: "reset", Local: empty, Remote: "/www/cache", Strategy: config.StrategyClean},
	}

	log := &recordingLogger{testLogger: testLogger{t}}
	if _, err := Run(context.Background(), cfg, log); err != nil {
		t.Fatal(err)
	}
	var warned bool
	for _, w := range log.warnings {
		if strings.Contains(w, "never re-uploads an unchanged file") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("expected the clean-inside-sync warning, got %v", log.warnings)
	}
}

// TestRunIsQuietWithoutTargetOverlaps makes sure the new warnings do not
// fire on an ordinary multi-deployment run.
func TestRunIsQuietWithoutTargetOverlaps(t *testing.T) {
	srv := startTestServer(t)
	site := t.TempDir()
	empty := t.TempDir()
	writeTree(t, site, map[string]string{"index.html": "site"})

	cfg := baseConfig(srv)
	cfg.Uploads = []config.UploadPair{
		{Name: "site", Local: site, Remote: "/www", Strategy: config.StrategySync},
		{Name: "reset", Local: empty, Remote: "/fresh", Strategy: config.StrategyClean},
	}

	log := &recordingLogger{testLogger: testLogger{t}}
	if _, err := Run(context.Background(), cfg, log); err != nil {
		t.Fatal(err)
	}
	for _, w := range log.warnings {
		if strings.Contains(w, "issue #278") {
			t.Errorf("disjoint targets should not warn, got %q", w)
		}
	}
}
