package benchmark_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The guards of issue #245, the corpus hygiene findings. Each one pins
// something that is true on this branch, was false on main, and can silently
// become false again: a gallery file force-added past its own ignore rule, the
// runner's instance name read back out of the environment, a trend file that
// mixes machines with nobody saying so, and a tagline that claims speed the
// corpus never measured. Like the fixture tests in schema/, these read the
// repository's own committed files, because those files are the contract.

// repoRoot is the checkout root, two levels up from this package.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("locating the repository root: %v", err)
	}
	return root
}

// repoFile reads a committed file and normalizes its line endings, so the
// assertions below can quote phrases the way they read in the editor.
func repoFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(name)))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}

// gitOutput runs git in the checkout and returns its stdout lines.
func gitOutput(t *testing.T, args ...string) []string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", repoRoot(t)}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	var lines []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimRight(line, "\r"); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// gitIgnores reports whether git's ignore rules match the path. The .git
// directory is what makes that a question about committed files rather than
// about whatever a local analysis run has left in out/ this week.
func gitIgnores(t *testing.T, path string) bool {
	t.Helper()
	err := exec.Command("git", "-C", repoRoot(t), "check-ignore", "-q", "--", path).Run()
	if err == nil {
		return true
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false
	}
	t.Fatalf("git check-ignore %s: %v", path, err)
	return false
}

// galleryNegations returns the files .gitignore un-ignores by name under
// benchmarks/analysis/out/, as forward-slash paths from the checkout root.
func galleryNegations(t *testing.T) []string {
	t.Helper()
	var names []string
	for _, raw := range strings.Split(string(repoFile(t, ".gitignore")), "\n") {
		if strings.HasPrefix(raw, "!/benchmarks/analysis/out/") {
			names = append(names, strings.TrimPrefix(strings.TrimPrefix(raw, "!"), "/"))
		}
	}
	return names
}

// galleryReadMeRefs returns the PNGs benchmarks/analysis/README.md embeds,
// in the same forward-slash form.
func galleryReadMeRefs(t *testing.T) []string {
	t.Helper()
	var names []string
	for _, raw := range strings.Split(string(repoFile(t, "benchmarks/analysis/README.md")), "\n") {
		for {
			i := strings.Index(raw, "](out/")
			if i < 0 {
				break
			}
			rest := raw[i+len("](out/"):]
			j := strings.Index(rest, ")")
			if j < 0 {
				break
			}
			if name := rest[:j]; strings.HasSuffix(name, ".png") {
				names = append(names, filepath.ToSlash(filepath.Join("benchmarks/analysis/out", name)))
			}
			raw = rest[j:]
		}
	}
	return names
}

// assertSameSet fails with the names only one side has. A plain length check
// would say nothing about which file dropped out, and that is the whole
// message a maintainer needs when the test fires.
func assertSameSet(t *testing.T, whatA string, a []string, whatB string, b []string) {
	t.Helper()
	left := append([]string(nil), a...)
	right := append([]string(nil), b...)
	sort.Strings(left)
	sort.Strings(right)
	for len(left) > 0 && len(right) > 0 && left[0] == right[0] {
		left, right = left[1:], right[1:]
	}
	for _, name := range left {
		t.Errorf("%s is %s but not %s (issue #245: adding a gallery file is a two-file change, the .gitignore negation and the reference on the analysis page)", name, whatA, whatB)
	}
	for _, name := range right {
		t.Errorf("%s is %s but not %s (issue #245: dropping a gallery file means dropping both halves)", name, whatB, whatA)
	}
}

// TestGalleryIsTrackedNamedAndReferredTo pins all three halves of the gallery
// mechanism of issue #245 item 3: the files git tracks under out/, the
// exceptions .gitignore names, and the images the analysis page embeds are the
// same set. The force-add this replaces was invisible: a file added with
// git add -f stays tracked while the ignore rule keeps saying out/ is
// ignored, and the next regeneration quietly does not update the gallery.
func TestGalleryIsTrackedNamedAndReferredTo(t *testing.T) {
	tracked := gitOutput(t, "ls-files", "--", "benchmarks/analysis/out/")
	negated := galleryNegations(t)
	referred := galleryReadMeRefs(t)
	if len(tracked) == 0 {
		t.Fatal("no files are tracked under benchmarks/analysis/out/; the gallery this test guards is gone")
	}
	assertSameSet(t, "tracked in git", tracked, "un-ignored in .gitignore", negated)
	assertSameSet(t, "un-ignored in .gitignore", negated, "embedded by benchmarks/analysis/README.md", referred)
	for _, name := range negated {
		if _, err := os.Stat(filepath.Join(repoRoot(t), filepath.FromSlash(name))); err != nil {
			t.Errorf("%s is un-ignored but not present in the checkout: %v", name, err)
		}
	}
}

// TestTheGalleryMechanismStillUnignores guards the shape of the ignore rule
// itself. Git cannot re-include a file whose parent directory is excluded, so
// the rule must ignore the contents of out/ (out/*) and not the directory
// (out/); with the directory form every negation below it silently stops
// working, the tracked files stay tracked, and the breakage only shows up
// when the next regeneration fails to update the gallery. Everything the
// exceptions do not name must stay ignored.
func TestTheGalleryMechanismStillUnignores(t *testing.T) {
	for _, name := range galleryNegations(t) {
		if gitIgnores(t, name) {
			t.Errorf("%s is named in .gitignore but still ignored; the out/ rule must match the directory's contents (out/*) and not the directory itself, or no negation under it works (issue #245)", name)
		}
	}
	if !gitIgnores(t, "benchmarks/analysis/out/not-in-the-gallery.png") {
		t.Error("a plot that no exception names is not ignored; out/ must stay ignored for everything but the gallery (issue #245)")
	}
}

// TestTheRunnerInstanceNameIsNotRead guards issue #245 item 2: the runner's
// instance name is infrastructure naming (a stored coolify-runner-O9nmBKiUox4Rm
// is what made the issue) with no analytical value that the environment block
// does not already carry, so nothing may read the variable again. The schema
// field stays, and the fixture tests in schema/ pin that stored results up to
// v3.8.3 still decode strictly and survive a round trip with it.
func TestTheRunnerInstanceNameIsNotRead(t *testing.T) {
	var hits []string
	for _, dir := range []string{"cmd", "internal", "scripts"} {
		err := filepath.WalkDir(filepath.Join(repoRoot(t), dir), func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			// The guard files themselves quote the literal, and a test is
			// not a place the variable can leak into a stored result from;
			// the production paths are what the scan is for.
			if !(strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".sh") || strings.HasSuffix(path, ".py")) {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if bytes.Contains(data, []byte("RUNNER_NAME")) {
				hits = append(hits, filepath.ToSlash(strings.TrimPrefix(path, repoRoot(t)+string(os.PathSeparator))))
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", dir, err)
		}
	}
	for _, hit := range hits {
		t.Errorf("%s reads the runner's instance name; it is infrastructure naming with no analytical value, and it used to be committed to every stored result (issue #245)", hit)
	}
}

// TestTrendRowsCarryTheRunnerWarning guards issue #245 item 1, the half the
// re-check left standing: trend.csv has carried a runner column since #194,
// but the pre-self-hosted rows still share the file and the median_ms column
// with everything later, so a reader who ignores the column sees a 5x step
// between v3.3.1 and v3.3.2 that is entirely a change of machine. While the
// file mixes runners, benchmarks/README.md must say so.
func TestTrendRowsCarryTheRunnerWarning(t *testing.T) {
	f, err := os.Open(filepath.Join(repoRoot(t), "benchmarks", "trend.csv"))
	if err != nil {
		t.Fatalf("opening trend.csv: %v", err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("reading trend.csv: %v", err)
	}
	if len(rows) < 2 {
		t.Fatal("trend.csv has no rows; the fixture this test reads is gone")
	}
	col := -1
	for i, name := range rows[0] {
		if name == "runner" {
			col = i
		}
	}
	if col < 0 {
		t.Fatalf("trend.csv has no runner column; it has been the comparability key since #194 and the whole warning rests on it")
	}
	runners := map[string]bool{}
	for _, row := range rows[1:] {
		if col < len(row) {
			runners[row[col]] = true
		}
	}
	if len(runners) < 2 {
		return // one environment in the file; there is nothing to warn about
	}
	warning := "Filter or group by `runner` before comparing"
	if !strings.Contains(string(repoFile(t, "benchmarks/README.md")), warning) {
		t.Errorf("trend.csv mixes %d runner environments, and benchmarks/README.md no longer warns that a plot must filter or group by the runner key before comparing rows (issue #245)", len(runners))
	}
}

// TestTheTaglineClaimsNoUnmeasuredSpeed guards issue #245 item 4: the corpus
// measures easySFTP against its own earlier releases and against the
// single-stream control of the line it ran over; no competing action has ever
// been measured, so the surfaces a reader sees first may not claim speed next
// to a table of competitors. If the claim is ever wanted, the honest order is
// to measure one competitor first, and the issue says the harness can already
// drive an arbitrary build.
func TestTheTaglineClaimsNoUnmeasuredSpeed(t *testing.T) {
	var tagline string
	for _, raw := range strings.Split(string(repoFile(t, "README.md")), "\n") {
		if strings.HasPrefix(raw, "**") && strings.HasSuffix(raw, "**") {
			tagline = raw
			break
		}
	}
	if tagline == "" {
		t.Fatal("README.md has no bold tagline line; the surface this test guards has moved")
	}
	if strings.Contains(strings.ToLower(tagline), "fast") {
		t.Errorf("the README tagline %q claims speed; the corpus has no comparative measurement behind it, only easySFTP against itself (issue #245)", tagline)
	}

	var action struct {
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal(repoFile(t, "action.yml"), &action); err != nil {
		t.Fatalf("parsing action.yml: %v", err)
	}
	if action.Description == "" {
		t.Fatal("action.yml has no description; the marketplace surface this test guards has moved")
	}
	if strings.Contains(strings.ToLower(action.Description), "fast") {
		t.Errorf("the action.yml description %q claims speed; the corpus has no comparative measurement behind it (issue #245)", action.Description)
	}

	site := string(repoFile(t, "site/index.html"))
	meta := between(site, `name="description" content="`, `"`)
	lede := between(site, `<p class="lede">`, `</p>`)
	for _, tc := range []struct{ what, text string }{
		{"the site's meta description", meta},
		{"the site's lede", lede},
	} {
		if tc.text == "" {
			t.Fatalf("site/index.html has no %s; the surface this test guards has moved", tc.what)
		}
		if strings.Contains(strings.ToLower(tc.text), "fast") {
			t.Errorf("%s (%q) claims speed; the corpus has no comparative measurement behind it (issue #245)", tc.what, tc.text)
		}
	}

	// While the README lists alternative actions, it must also carry the note
	// that says what the corpus does and does not compare. The phrase wraps
	// across lines on the page, so the check folds whitespace before it
	// compares.
	readme := oneline(string(repoFile(t, "README.md")))
	if strings.Contains(readme, "wlixcc-rel") &&
		!strings.Contains(readme, oneline("compares easySFTP to another action")) {
		t.Error("README.md lists alternative actions but no longer carries the note that the benchmark corpus compares easySFTP to itself, not to them (issue #245)")
	}
}

// oneline folds every run of whitespace into a single space, so a phrase that
// wraps across a line in a Markdown file still matches.
func oneline(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// between returns the text from the first occurrence of open to the next
// close, or the empty string when either is missing.
func between(s, open, close string) string {
	i := strings.Index(s, open)
	if i < 0 {
		return ""
	}
	rest := s[i+len(open):]
	j := strings.Index(rest, close)
	if j < 0 {
		return ""
	}
	return rest[:j]
}
