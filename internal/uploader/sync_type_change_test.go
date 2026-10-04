package uploader

import (
	"context"
	"strings"
	"testing"

	"github.com/eiserv/easySFTP/internal/config"
)

// TestSyncReplacesAFileWithADirectory pins the file-to-directory type change:
// the last sync uploaded "about" as a file, the build now produces
// "about/index.html". The stale manifest entry is removed before the upload,
// so the directory can be created; the entry counts as deleted, the new file
// as uploaded, and a second run is a no-op. Without the pre-pass the upload
// fails on MkdirAll over the file and every retry fails identically
// (issue #280).
func TestSyncReplacesAFileWithADirectory(t *testing.T) {
	srv := startTestServer(t)
	fileVersion := t.TempDir()
	writeTree(t, fileVersion, map[string]string{"index.html": "home", "about": "about page"})
	dirVersion := t.TempDir()
	writeTree(t, dirVersion, map[string]string{"index.html": "home", "about/index.html": "about page"})

	cfg := baseConfig(srv)
	pair := config.UploadPair{Name: "site", Local: fileVersion, Remote: "/site", Strategy: config.StrategySync}
	cfg.Uploads = []config.UploadPair{pair}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	pair.Local = dirVersion
	cfg.Uploads = []config.UploadPair{pair}
	stats, err := Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatalf("the file-to-directory change must deploy, got: %v", err)
	}
	if stats.FilesDeleted != 1 || stats.FilesUploaded != 1 {
		t.Fatalf("run 2: deleted %d (want 1, the stale file), uploaded %d (want 1, the new page)",
			stats.FilesDeleted, stats.FilesUploaded)
	}
	if !remoteExists(t, srv, "/site/about/index.html") {
		t.Error("run 2: the new page must be on the server")
	}
	if remoteExists(t, srv, "/site/about") && !remoteExists(t, srv, "/site/about/index.html") {
		t.Error("run 2: the stale file must not survive as a path blocker")
	}

	// The retry is a no-op: the manifest matches the tree.
	stats, err = Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesUploaded != 0 || stats.FilesDeleted != 0 {
		t.Errorf("run 3: a settled sync should change nothing, uploaded %d deleted %d",
			stats.FilesUploaded, stats.FilesDeleted)
	}
}

// TestSyncReplacesADirectoryWithAFile pins the reverse type change: the last
// sync uploaded "about/index.html", the build now produces "about" as a
// single file. The stale entry goes before the upload and the emptied
// directory is removed so the rename does not meet "is a directory" (issue
// #280).
func TestSyncReplacesADirectoryWithAFile(t *testing.T) {
	srv := startTestServer(t)
	dirVersion := t.TempDir()
	writeTree(t, dirVersion, map[string]string{"index.html": "home", "about/index.html": "about page"})
	fileVersion := t.TempDir()
	writeTree(t, fileVersion, map[string]string{"index.html": "home", "about": "now a file"})

	cfg := baseConfig(srv)
	pair := config.UploadPair{Name: "site", Local: dirVersion, Remote: "/site", Strategy: config.StrategySync}
	cfg.Uploads = []config.UploadPair{pair}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	pair.Local = fileVersion
	cfg.Uploads = []config.UploadPair{pair}
	stats, err := Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatalf("the directory-to-file change must deploy, got: %v", err)
	}
	if stats.FilesDeleted != 1 || stats.FilesUploaded != 1 {
		t.Fatalf("run 2: deleted %d (want 1, the stale page), uploaded %d (want 1, the new file)",
			stats.FilesDeleted, stats.FilesUploaded)
	}
	if !remoteExists(t, srv, "/site/about") {
		t.Error("run 2: the replacement file must be on the server")
	}

	stats, err = Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesUploaded != 0 || stats.FilesDeleted != 0 {
		t.Errorf("run 3: a settled sync should change nothing, uploaded %d deleted %d",
			stats.FilesUploaded, stats.FilesDeleted)
	}
}

// TestSyncRefusesAFileOverAnUnmanagedDirectory pins the guard the issue asks
// for: a directory that still holds files this deployment did not upload is
// not deleted to make room. The run fails, but with a message that says what
// is in the way and what to do, instead of the bare "is a directory" from
// the rename.
func TestSyncRefusesAFileOverAnUnmanagedDirectory(t *testing.T) {
	srv := startTestServer(t)
	dirVersion := t.TempDir()
	writeTree(t, dirVersion, map[string]string{"index.html": "home", "about/index.html": "about page"})
	fileVersion := t.TempDir()
	writeTree(t, fileVersion, map[string]string{"index.html": "home", "about": "now a file"})

	cfg := baseConfig(srv)
	pair := config.UploadPair{Name: "site", Local: dirVersion, Remote: "/site", Strategy: config.StrategySync}
	cfg.Uploads = []config.UploadPair{pair}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	// Someone else puts a file into the synced directory that this
	// deployment never uploaded. The stale manifest entry for the owned
	// page still goes, but the unmanaged one must block the type change
	// rather than be deleted to make room.
	client := srv.verifyClient(t)
	f, err := client.Create("/site/about/unmanaged.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("out of band")); err != nil {
		t.Fatal(err)
	}
	f.Close()

	pair.Local = fileVersion
	cfg.Uploads = []config.UploadPair{pair}
	_, runErr := Run(context.Background(), cfg, testLogger{t})
	if runErr == nil {
		t.Fatal("the run must refuse to delete a directory it does not own the contents of")
	}
	if !strings.Contains(runErr.Error(), "this deployment did not upload") {
		t.Errorf("the error must say the directory holds unmanaged entries, got: %v", runErr)
	}
	if !strings.Contains(runErr.Error(), "mode: clean") {
		t.Errorf("the error must point at mode: clean as the way out, got: %v", runErr)
	}
	// The unmanaged content is still there.
	if !remoteExists(t, srv, "/site/about/unmanaged.txt") {
		t.Error("the unmanaged file must survive the refused run")
	}
}

// TestSyncTypeChangeRespectsMaxDeletes pins the budget interaction: the
// colliding file removals are part of the reservation the run already made,
// but the directories they empty are charged one at a time, before the
// attempt, like every removal whose count could not be known in advance. A
// spent budget refuses the type change with the reason instead of letting
// the upload fail on the un-removed directory with a bare rename error.
func TestSyncTypeChangeRespectsMaxDeletes(t *testing.T) {
	srv := startTestServer(t)
	dirVersion := t.TempDir()
	writeTree(t, dirVersion, map[string]string{"index.html": "home", "about/index.html": "about page"})
	fileVersion := t.TempDir()
	writeTree(t, fileVersion, map[string]string{"index.html": "home", "about": "now a file"})

	cfg := baseConfig(srv)
	pair := config.UploadPair{Name: "site", Local: dirVersion, Remote: "/site", Strategy: config.StrategySync}
	cfg.Uploads = []config.UploadPair{pair}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	// One delete covers the stale page; the emptied directory is one more.
	cfg.Safety.MaxDeletes = 1
	pair.Local = fileVersion
	cfg.Uploads = []config.UploadPair{pair}
	_, err := Run(context.Background(), cfg, testLogger{t})
	if err == nil {
		t.Fatal("the run must refuse the type change that would exceed safety.max_deletes")
	}
	if !strings.Contains(err.Error(), "safety.max_deletes") {
		t.Errorf("the error must name the budget, got: %v", err)
	}
	// The stale page is gone; the directory, refused, is still there.
	if remoteExists(t, srv, "/site/about/index.html") {
		t.Error("the reserved file removal must have run")
	}
	if !remoteExists(t, srv, "/site/about") {
		t.Error("the directory the budget refused must still be there")
	}
}

// TestSyncClearsTwoStaleFilesInOneChain pins the multi-level case the
// ancestor walk in collidingDeletes promises: a manifest holding stale
// files at two levels of one chain ("about" and "about/team") must lose
// both when the plan puts files below them, not just the deepest one.
func TestSyncClearsTwoStaleFilesInOneChain(t *testing.T) {
	srv := startTestServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"index.html": "home", "about/team/index.html": "team"})

	// A previous sync is simulated by hand: the manifest lists two stale
	// files in one chain, "about" and "about/team", but the server now
	// holds "about" as a single file (someone flattened the directory,
	// the hand-mangled state). "about/team" exists only as a manifest
	// entry: the pre-pass must take the real file away for the plan's
	// directories to be creatable, and the phantom entry must be dropped
	// from the manifest too, by the same already-gone rule the delete
	// sweep applies.
	seedRemoteFile(t, srv, "/www", manifestName, `{
	  "version": 3,
	  "files": {
	    "about": {"hash": "a1", "size": 4, "mtime": 1},
	    "about/team": {"hash": "b2", "size": 4, "mtime": 1}
	  }
	}`)
	seedRemoteFile(t, srv, "/www", "about", "stale")

	stats, err := Run(context.Background(), syncConfig(srv, local), testLogger{t})
	if err != nil {
		t.Fatalf("a two-level type change must deploy, got: %v", err)
	}
	if stats.FilesDeleted != 2 || stats.FilesUploaded != 2 {
		t.Fatalf("deleted %d (want 2, the real chain file and the phantom entry), uploaded %d (want 2)",
			stats.FilesDeleted, stats.FilesUploaded)
	}
	if !remoteExists(t, srv, "/www/about/team/index.html") {
		t.Error("the new nested page must be on the server")
	}
	// A settled second run changes nothing, so the manifest agrees with
	// the tree and the chain is really gone as files.
	stats, err = Run(context.Background(), syncConfig(srv, local), testLogger{t})
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesUploaded != 0 || stats.FilesDeleted != 0 {
		t.Errorf("run 2: a settled sync should change nothing, uploaded %d deleted %d",
			stats.FilesUploaded, stats.FilesDeleted)
	}
}

// TestSyncPrepassDoesNotTouchDisjointStaleFiles pins the narrowness of the
// pre-pass: a stale file whose path is neither a planned directory nor
// under a planned file is not moved ahead of the upload. It is deleted in
// the normal sweep, after the upload, keeping the upload-first order that
// leaves the old site serving until the new files are in place.
func TestSyncPrepassDoesNotTouchDisjointStaleFiles(t *testing.T) {
	srv := startTestServer(t)
	local := t.TempDir()
	writeTree(t, local, map[string]string{"index.html": "home", "about/index.html": "about page"})

	seedRemoteFile(t, srv, "/www", manifestName, `{
	  "version": 3,
	  "files": {
	    "old.html": {"hash": "a1", "size": 8, "mtime": 1},
	    "about": {"hash": "b2", "size": 4, "mtime": 1}
	  }
	}`)
	seedRemoteFile(t, srv, "/www", "old.html", "stale")
	seedRemoteFile(t, srv, "/www", "about", "stale")

	stats, err := Run(context.Background(), syncConfig(srv, local), testLogger{t})
	if err != nil {
		t.Fatalf("run 1: %v", err)
	}
	// The colliding entry went early; the disjoint stale file went in the
	// normal sweep: same total, and both are gone.
	if stats.FilesDeleted != 2 || stats.FilesUploaded != 2 {
		t.Fatalf("deleted %d (want 2), uploaded %d (want 2)",
			stats.FilesDeleted, stats.FilesUploaded)
	}
	if remoteExists(t, srv, "/www/old.html") {
		t.Error("the disjoint stale file must be gone after the sweep")
	}
	if !remoteExists(t, srv, "/www/about/index.html") {
		t.Error("the new page must be on the server")
	}
}

// TestSyncFileToDirectoryKeepsRetainedSiblings pins the narrowness the parent
// prune must have (review of #297): a file-to-directory transition replaces
// one manifest entry and has no planned file above it, so no directory is
// removed. The old parent walk collected "docs" anyway, saw the retained
// keep.html in its ReadDir, and aborted a valid deploy with the
// unmanaged-entries error -- a file the plan never asked to remove.
func TestSyncFileToDirectoryKeepsRetainedSiblings(t *testing.T) {
	srv := startTestServer(t)
	fileVersion := t.TempDir()
	writeTree(t, fileVersion, map[string]string{"index.html": "home", "docs/about": "about page", "docs/keep.html": "kept"})
	dirVersion := t.TempDir()
	writeTree(t, dirVersion, map[string]string{"index.html": "home", "docs/about/index.html": "about page", "docs/keep.html": "kept"})

	cfg := baseConfig(srv)
	pair := config.UploadPair{Name: "site", Local: fileVersion, Remote: "/site", Strategy: config.StrategySync}
	cfg.Uploads = []config.UploadPair{pair}
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	pair.Local = dirVersion
	cfg.Uploads = []config.UploadPair{pair}
	stats, err := Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatalf("the file-to-directory change under a retained sibling must deploy, got: %v", err)
	}
	if stats.FilesDeleted != 1 || stats.FilesUploaded != 1 {
		t.Fatalf("run 2: deleted %d (want 1, the stale file), uploaded %d (want 1, the new page)", stats.FilesDeleted, stats.FilesUploaded)
	}
	if stats.DirsDeleted != 0 {
		t.Errorf("run 2: a file-to-directory transition removes no directory, removed %d", stats.DirsDeleted)
	}
	if !remoteExists(t, srv, "/site/docs/about/index.html") {
		t.Error("run 2: the new nested page must be on the server")
	}
	if !remoteExists(t, srv, "/site/docs/keep.html") {
		t.Error("run 2: the retained sibling must survive the type change")
	}

	// The retry is a no-op: the manifest matches the tree.
	stats, err = Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesUploaded != 0 || stats.FilesDeleted != 0 {
		t.Errorf("run 3: a settled sync should change nothing, uploaded %d deleted %d", stats.FilesUploaded, stats.FilesDeleted)
	}
}

// TestSyncRetriesAnInterruptedDirectoryToFileChange pins the recovery contract
// (review of #297): when clearCollideParents fails after the colliding files
// are already gone, the recovery manifest must keep those entries, so the
// retry re-discovers the collision, re-enters the parent prune (the
// already-gone file counts as deleted), removes the leftover directory, and
// the type change completes. Dropping the pre-deleted entries stranded the
// retry: no stale child meant no colliding entry, no prune, and the upload
// failing on the leftover directory forever.
func TestSyncRetriesAnInterruptedDirectoryToFileChange(t *testing.T) {
	// The rmdir of /www/about fails once: the interrupted run's state is
	// exactly what a server-side rmdir refusal or a crashed run leaves.
	fault := &faultyPathCmd{method: "Rmdir", path: "/www/about"}
	srv := startTestServer(t, withFaultyPath(fault))

	dirVersion := t.TempDir()
	writeTree(t, dirVersion, map[string]string{"index.html": "home", "about/index.html": "about page"})
	fileVersion := t.TempDir()
	writeTree(t, fileVersion, map[string]string{"index.html": "home", "about": "now a file"})

	cfg := syncConfig(srv, dirVersion)
	if _, err := Run(context.Background(), cfg, testLogger{t}); err != nil {
		t.Fatal(err)
	}

	// The type-change run fails at the injected rmdir, after the colliding
	// page was already deleted.
	fault.enabled.Store(true)
	cfg = syncConfig(srv, fileVersion)
	_, runErr := Run(context.Background(), cfg, testLogger{t})
	if runErr == nil {
		t.Fatal("the run must fail while the rmdir is injected to fail")
	}
	if !strings.Contains(runErr.Error(), "removing remote directory") {
		t.Errorf("the error must be the directory removal, got: %v", runErr)
	}
	// The emptied directory is still in the way, and the manifest must still
	// know about the colliding entry, or the retry below has nothing to
	// re-discover.
	if !remoteExists(t, srv, "/www/about") {
		t.Fatal("the directory the failed rmdir left behind must still be there")
	}
	m := readRemoteManifest(t, srv)
	if _, ok := m.Files["about/index.html"]; !ok {
		t.Fatalf("the recovery manifest must retain the pre-deleted colliding entry for the retry to re-discover, got files %v", m.Files)
	}

	// The retry: same config, no injected failure. The retained entry
	// re-enters the pre-pass (already-gone counts as deleted), the prune
	// removes the leftover directory, and the change completes.
	fault.enabled.Store(false)
	stats, err := Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatalf("the retry must complete the interrupted type change, got: %v", err)
	}
	if stats.DirsDeleted != 1 {
		t.Errorf("the retry must remove the leftover directory, removed %d", stats.DirsDeleted)
	}
	if got := readRemote(t, srv, "/www/about"); got != "now a file" {
		t.Errorf("the replacement file must be on the server, got %q", got)
	}

	// Settled: the manifest matches the tree.
	stats, err = Run(context.Background(), cfg, testLogger{t})
	if err != nil {
		t.Fatal(err)
	}
	if stats.FilesUploaded != 0 || stats.FilesDeleted != 0 {
		t.Errorf("run 4: a settled sync should change nothing, uploaded %d deleted %d", stats.FilesUploaded, stats.FilesDeleted)
	}
}
