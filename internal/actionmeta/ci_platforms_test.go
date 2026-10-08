package actionmeta_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// These tests pin the two halves of the platform claim the README makes
// ("Linux / macOS / Windows: yes / yes / yes") against the repository that
// backs it (issue #238):
//
//   - the unit matrix in ci.yml builds and tests every advertised platform,
//     instead of shipping macOS binaries that no CI job ever compiled or ran;
//   - the launcher scripts stay clean of bash 4 constructs, because the
//     platform the action advertises is also the platform whose /bin/bash is
//     3.2. A self-hosted macOS runner without a Homebrew bash must be able to
//     run the launcher, not merely the GitHub-hosted one that happens to
//     have bash 4 early in PATH.

type ciWorkflow struct {
	Jobs map[string]struct {
		Strategy struct {
			Matrix struct {
				Os []string `yaml:"os"`
			} `yaml:"matrix"`
		} `yaml:"strategy"`
	} `yaml:"jobs"`
}

// advertisedPlatforms is the set the README's comparison table claims and
// release-binaries.yml publishes for. A platform added to either of those
// belongs here too.
var advertisedPlatforms = []string{"ubuntu-latest", "windows-latest", "macos-latest"}

func loadCIWorkflow(t *testing.T) ciWorkflow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("reading ci.yml: %v", err)
	}
	var wf ciWorkflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatalf("parsing ci.yml: %v", err)
	}
	return wf
}

// TestUnitMatrixCoversAdvertisedPlatforms: every job that compiles and tests
// the action's Go code (the launcher and metadata job, the unit test job)
// must sweep the full advertised platform set. The release workflow only
// smoke-tests the macOS binaries (--help), so this matrix is the only job
// that compiles and tests the code macOS users download.
func TestUnitMatrixCoversAdvertisedPlatforms(t *testing.T) {
	wf := loadCIWorkflow(t)
	for _, job := range []string{"action-tests", "test"} {
		j, ok := wf.Jobs[job]
		if !ok {
			t.Fatalf("ci.yml has no job %q; the unit matrix test no longer guards anything", job)
		}
		for _, want := range advertisedPlatforms {
			found := false
			for _, got := range j.Strategy.Matrix.Os {
				if got == want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("ci.yml: the %s job does not run on %s; the README advertises the platform and release-binaries.yml ships it, so the matrix must build and test it (issue #238)", job, want)
			}
		}
	}
}

// bash4OnlyConstructs are the constructs the launcher scripts must not use:
// mapfile/readarray and coproc (bash 4), associative arrays (4.0), the |& and
// &>> redirections (4.0/4.1) and the case-modifier expansions (4.0). macOS
// ships bash 3.2 as /bin/bash, and a self-hosted macOS runner without a
// Homebrew bash resolves exactly that one.
var bash4OnlyConstructs = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"mapfile", regexp.MustCompile(`(^|[^[:alnum:]_])mapfile[[:space:]]`)},
	{"readarray", regexp.MustCompile(`(^|[^[:alnum:]_])readarray[[:space:]]`)},
	{"coproc", regexp.MustCompile(`(^|[^[:alnum:]_])coproc[[:space:]]`)},
	{"associative array", regexp.MustCompile(`(^|[;&])[[:space:]]*(declare|local|typeset)[[:space:]]+[^\n]*-[[:alnum:]]*A`)},
	{"case-modifier expansion", regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*(\[[0-9@*]+\])?[,^~]`)},
	{"|& redirection", regexp.MustCompile(`[^|]\|&[[:space:]]`)},
	{"&>> redirection", regexp.MustCompile(`&>>`)},
}

// TestLauncherScriptsStayBash32Clean: scripts/*.sh run on the customer's
// runner, whatever bash it resolves. The constructs above work on the
// GitHub-hosted macOS image (bash 4+ in PATH) and fail on /bin/bash 3.2, so
// a script that green-lights CI can still break a self-hosted macOS user,
// the same class of gap as #62, caught only by the customer.
func TestLauncherScriptsStayBash32Clean(t *testing.T) {
	scripts, err := filepath.Glob(filepath.Join("..", "..", "scripts", "*.sh"))
	if err != nil {
		t.Fatalf("listing scripts: %v", err)
	}
	if len(scripts) == 0 {
		t.Fatal("no scripts found; the test no longer guards anything")
	}
	for _, path := range scripts {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		name := filepath.Base(path)
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "#") {
				continue // comments may name the constructs we forbid
			}
			for _, c := range bash4OnlyConstructs {
				if c.pattern.MatchString(line) {
					t.Errorf("%s: uses the bash 4 %s, which /bin/bash 3.2 on macOS does not support: %s", name, c.name, trimmed)
				}
			}
		}
	}
}
