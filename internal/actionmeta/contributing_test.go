package actionmeta_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/eiserv/easySFTP/internal/config"
	"gopkg.in/yaml.v3"
)

// CONTRIBUTING.md is the first document a contributor opens, and the only
// one that shows a complete EASYSFTP_* environment: the command that runs
// the binary locally, and the map of the repository. Both went stale in
// ways that break a newcomer's first command (issue #232): the example set
// EASYSFTP_SERVER and EASYSFTP_UPLOADS, two inputs v3 removed, so the
// documented run died in checkRemovedInputs with a migration error that
// reads like the contributor's own environment is wrong; and the layout
// map covered the action core but hid the benchmark harness and half the
// internal packages, together the majority of the Go code in this
// repository.
//
// These tests keep the guide honest the same way TestActionMetadata keeps
// action.yml honest: every EASYSFTP_* variable the guide names has to be a
// live input (a removed one does not misbehave, it fails the run before
// anything happens), the run example has to load through the real config
// parser, and the layout map has to name every binary, internal package
// and top-level directory that exists, and nothing that does not.

// readAction parses action.yml with the same struct TestActionMetadata
// uses, so these tests and the wiring test can never disagree about the
// input set.
func readAction(t *testing.T) metadata {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var action metadata
	if err := yaml.Unmarshal(data, &action); err != nil {
		t.Fatalf("parse action.yml: %v", err)
	}
	return action
}

// readGuide returns the CONTRIBUTING.md text.
func readGuide(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "CONTRIBUTING.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// fencedBlock returns the body of the first fenced code block after the
// given heading. The layout map and the run example are the two blocks
// these tests consume.
func fencedBlock(t *testing.T, guide, heading string) string {
	t.Helper()
	at := strings.Index(guide, heading)
	if at < 0 {
		t.Fatalf("CONTRIBUTING.md has no %q heading", heading)
	}
	rest := guide[at:]
	loc := regexp.MustCompile("\n```[a-z]*\n").FindStringIndex(rest)
	if loc == nil {
		t.Fatalf("no code block after %q", heading)
	}
	body := rest[loc[1]:]
	end := strings.Index(body, "\n```")
	if end < 0 {
		t.Fatalf("the code block after %q is not closed", heading)
	}
	return body[:end]
}

// inputSplit divides action.yml's inputs into the live v3 inputs and the
// removed-input tombstones, both keyed by input name. An input is a
// tombstone when its own description says it was removed: that marker is
// what the migration documentation quotes, so it is the derivation these
// tests use too.
func inputSplit(action metadata) (live, tombstones map[string]string) {
	live = map[string]string{}
	tombstones = map[string]string{}
	for name, input := range action.Inputs {
		if strings.Contains(input.Description, "Removed in v") {
			tombstones[name] = input.Description
		} else {
			live[name] = input.Description
		}
	}
	return live, tombstones
}

// envToken matches one EASYSFTP_* variable name. The bare EASYSFTP_* the
// guide uses to speak about the environment as a whole does not match,
// because the class requires at least one character and excludes "*".
var envToken = regexp.MustCompile(`EASYSFTP_[A-Z0-9_]+`)

// inputName converts an EASYSFTP_* variable to its action.yml input name.
func inputName(env string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(env, "EASYSFTP_")), "_", "-")
}

// TestContributingNamesOnlyLiveInputs fails the moment the guide names a
// removed v2 input instead of its live replacement: a removed input does
// not misbehave, it fails the run with a migration error before anything
// happens, so the documented command cannot even reach the connect stage.
// This is the regression behind issue #232's first finding.
func TestContributingNamesOnlyLiveInputs(t *testing.T) {
	action := readAction(t)
	guide := readGuide(t)
	live, tombstones := inputSplit(action)

	for _, token := range envToken.FindAllString(guide, -1) {
		name := inputName(token)
		if hint, dead := tombstones[name]; dead {
			t.Errorf("the guide names %s, an input the action no longer accepts (%s), so the documented command fails the run with a migration error", token, hint)
			continue
		}
		if _, ok := live[name]; !ok {
			t.Errorf("the guide names %s, which is not a declared action input", token)
		}
	}
}

// TestTombstoneInputsMatchConfig pins the two copies of the removed-input
// list against each other: the input declarations in action.yml (what a
// workflow author reads) and removedInputs in internal/config/config.go
// (what actually fails the run). Neither direction of drift is visible on
// its own: an input declared as removed but not enforced is silently
// ignored, and an enforced removal with no input left declared produces an
// error naming an input the action does not document.
func TestTombstoneInputsMatchConfig(t *testing.T) {
	action := readAction(t)
	_, tombstones := inputSplit(action)

	declared := map[string]bool{}
	for name := range tombstones {
		env := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
		// build-mode never reaches the Go binary: the runner hands it to
		// scripts/prepare-action.sh as INPUT_BUILD_MODE, which fails the
		// run before the upload step starts.
		if env == "BUILD_MODE" {
			continue
		}
		declared[env] = true
	}

	enforced := removedInputsFromConfig(t)
	for env := range enforced {
		if !declared[env] {
			t.Errorf("internal/config rejects EASYSFTP_%s as a removed input, but no action.yml input declares itself removed", env)
		}
	}
	for env := range declared {
		if !enforced[env] {
			t.Errorf("action.yml declares the %s input removed, but internal/config does not name EASYSFTP_%s in removedInputs; the upload step still wires it, so a workflow passing it is silently ignored", strings.ToLower(strings.ReplaceAll(env, "_", "-")), env)
		}
	}
}

// removedInputsFromConfig reads the removedInputs literal out of
// internal/config/config.go: the copy of the removed-input list that
// enforces the tombstones at run time. Parsing the source pins the real
// list instead of restating it a third time here, the same way
// TestActionMetadata parses action.yml rather than restating its wiring.
func removedInputsFromConfig(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "config", "config.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	at := strings.Index(source, "var removedInputs = []removedInput{")
	if at < 0 {
		t.Fatal("internal/config/config.go no longer declares removedInputs")
	}
	end := strings.Index(source[at:], "\n}")
	if end < 0 {
		t.Fatal("cannot find the end of the removedInputs literal")
	}
	enforced := map[string]bool{}
	for _, m := range regexp.MustCompile(`\{"([A-Z_]+)",`).FindAllStringSubmatch(source[at:at+end], -1) {
		enforced[m[1]] = true
	}
	if len(enforced) == 0 {
		t.Fatal("the removedInputs parse found no entries; the literal format changed")
	}
	return enforced
}

// uploadStepOf returns the step that runs the binary, whose env block is
// the complete set of EASYSFTP_* variables the runner can hand the
// process.
func uploadStepOf(t *testing.T, action metadata) *actionStep {
	t.Helper()
	for i, step := range action.Runs.Steps {
		if step.Name == "Upload via SFTP" {
			return &action.Runs.Steps[i]
		}
	}
	t.Fatal("step \"Upload via SFTP\" is missing")
	return nil
}

// TestContributingExampleLoadsThroughConfig runs the guide's local-run
// example through config.Load, the same parser the documented command
// uses. The guide is then a tested artifact like docs/easysftp.example.yml
// (see internal/config/schema_parity_test.go): if the example names a
// removed input, misstates a value the load drops, or stops being the
// documented entry point, the failure is here rather than in a newcomer's
// terminal.
func TestContributingExampleLoadsThroughConfig(t *testing.T) {
	action := readAction(t)
	guide := readGuide(t)
	block := fencedBlock(t, guide, "### Running the binary locally")

	// The block is a shell session: join the continuation lines and split
	// it into environment assignments plus the command.
	var fields []string
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "$ "))
		line = strings.TrimSpace(strings.TrimSuffix(line, "\\"))
		fields = append(fields, strings.Fields(line)...)
	}
	if len(fields) < 4 || !slices.Equal(fields[len(fields)-3:], []string{"go", "run", "./cmd/easysftp"}) {
		t.Fatalf("the guide's run example should end with `go run ./cmd/easysftp`, got %q", strings.Join(fields, " "))
	}

	// Start from the environment the runner hands the binary: every
	// EASYSFTP_* variable action.yml wires is unset, then the example's
	// assignments are applied on top, so ambient EASYSFTP_* values in a
	// developer's shell cannot change what is being tested.
	upload := uploadStepOf(t, action)
	for env := range upload.Env {
		t.Setenv(env, "")
	}
	for _, assignment := range fields[:len(fields)-3] {
		name, value, ok := strings.Cut(assignment, "=")
		if !ok {
			t.Fatalf("the guide's run example has %q where NAME=value belongs", assignment)
		}
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		t.Setenv(name, value)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("the guide's run example must load through the real config parser: %v", err)
	}
	if cfg.Server != "localhost" || cfg.Port != 2222 || cfg.Username != "demo" || cfg.Password == "" {
		t.Errorf("connection settings from the example: server=%q port=%d username=%q password set=%v", cfg.Server, cfg.Port, cfg.Username, cfg.Password != "")
	}
	if !cfg.AllowAnyHostKey {
		t.Error("the example sets allow-any-host-key for the throwaway local server; the load must carry it")
	}
	if !cfg.DryRun {
		t.Error("the example sets dry-run; the load must carry it")
	}
	if len(cfg.Uploads) != 1 || cfg.Uploads[0].Local != "./dist" || cfg.Uploads[0].Remote != "/upload" || cfg.Uploads[0].Strategy != config.StrategyOverlay {
		t.Errorf("deployment from the example: %+v", cfg.Uploads)
	}
}

// subPackageClaim matches a "(N sub-packages: a, b, ...)" claim in the
// layout map, once whitespace is flattened.
var subPackageClaim = regexp.MustCompile(`\((\d+) sub-packages: ([a-z, ]+)\)`)

func TestContributingRepositoryLayout(t *testing.T) {
	guide := readGuide(t)
	block := fencedBlock(t, guide, "### Repository layout")

	// Every line that starts at the left margin names a path; the
	// indented lines behind it are the description.
	listed := map[string]bool{}
	for _, line := range strings.Split(block, "\n") {
		if len(line) == 0 || line[0] == ' ' || line[0] == '\t' {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		listed[fields[0]] = true
	}

	root := filepath.Join("..", "..")

	// Nothing the map names may go stale.
	for path := range listed {
		if _, err := os.Stat(filepath.Join(root, strings.TrimSuffix(path, "/"))); err != nil {
			t.Errorf("the layout map names %s, which does not exist", path)
		}
	}

	// And the map may not hide parts of the repository again: every
	// binary, internal package and top-level directory must be in it
	// (issue #232).
	want := []string{"action.yml", "schema/", "benchmarks/", "scripts/", "site/", "docs/"}
	for _, dir := range dirsUnder(t, filepath.Join(root, "cmd")) {
		want = append(want, "cmd/"+dir+"/")
	}
	for _, dir := range dirsUnder(t, filepath.Join(root, "internal")) {
		want = append(want, "internal/"+dir+"/")
	}
	for _, path := range want {
		if !listed[path] {
			t.Errorf("the layout map does not mention %s; a newcomer reading it concludes that part of the repository does not exist (issue #232)", path)
		}
	}

	// If the map states how many sub-packages the benchmark harness has,
	// the statement has to stay true.
	flat := regexp.MustCompile(`\s+`).ReplaceAllString(block, " ")
	if m := subPackageClaim.FindStringSubmatch(flat); m != nil {
		names := dirsUnder(t, filepath.Join(root, "internal", "benchmark"))
		claimed := []string{}
		for _, name := range strings.Split(m[2], ",") {
			if name = strings.TrimSpace(name); name != "" {
				claimed = append(claimed, name)
			}
		}
		slices.Sort(names)
		slices.Sort(claimed)
		if m[1] != strconv.Itoa(len(names)) || !slices.Equal(names, claimed) {
			t.Errorf("the layout map says internal/benchmark has %s sub-packages (%s); the repository has %d (%s)", m[1], m[2], len(names), strings.Join(names, ", "))
		}
	}
}

// dirsUnder returns the names of the directories directly under dir,
// testdata excluded: it is a fixture directory, not a package.
func dirsUnder(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && e.Name() != "testdata" {
			names = append(names, e.Name())
		}
	}
	return names
}
