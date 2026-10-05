package report_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eiserv/easySFTP/internal/benchmark/report"
	"github.com/eiserv/easySFTP/internal/benchmark/schema"
)

// The committed results under benchmarks/ are the goldens this renderer is
// responsible for: internal/benchmark/output.go writes the CSV from exactly
// the Standard and Matrix values below, the store copies it verbatim into the
// repository, and the never-rewrite rule means a rendering regression there is
// permanent the moment it lands (issue #239).
//
// Re-decoding every stored JSON and re-rendering its CSV compares the renderer
// against the corpus itself, so column order, number formatting and the
// null-as-empty-field convention are all pinned by data that already exists
// rather than by a fixture written to match the code.

// corpusCSVs lists every stored CSV that today's renderer shape can reproduce.
// The three under archive/matrix predate the current column set (they carry no
// link_profile, rtt_p50_ms or control_single_mib_per_s and no
// request_concurrency_used), so they are not goldens for the current renderer;
// the fixture test in internal/benchmark/schema keeps them readable instead.
func corpusCSVs(t *testing.T) []string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "benchmarks"))
	if err != nil {
		t.Fatalf("locating benchmarks/: %v", err)
	}
	var paths []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".csv" {
			return nil
		}
		if filepath.Base(path) == "trend.csv" {
			// Written from index.go's own column list, not by this package.
			return nil
		}
		if strings.Contains(filepath.ToSlash(path), "/archive/") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walking benchmarks/: %v", err)
	}
	if len(paths) == 0 {
		t.Fatal("no stored CSVs found; the corpus test would pass vacuously")
	}
	return paths
}

func TestStoredResultsRenderTheirCommittedCSV(t *testing.T) {
	for _, csvPath := range corpusCSVs(t) {
		t.Run(corpusName(t, csvPath), func(t *testing.T) {
			data, err := os.ReadFile(strings.TrimSuffix(csvPath, ".csv") + ".json")
			if err != nil {
				t.Fatalf("reading the stored measurement: %v", err)
			}
			env, err := schema.DecodeStrict[schema.Envelope](data)
			if err != nil {
				t.Fatalf("decoding the envelope: %v", err)
			}
			stored, err := os.ReadFile(csvPath)
			if err != nil {
				t.Fatalf("reading the stored CSV: %v", err)
			}
			// The index is LF and the working tree may be CRLF; the renderer
			// writes LF, so the comparison is on content and not on checkout
			// configuration.
			want := strings.ReplaceAll(string(stored), "\r\n", "\n")

			kind, err := env.BenchmarkKind()
			if err != nil {
				t.Fatalf("reading the benchmark kind: %v", err)
			}
			var got string
			switch kind {
			case schema.BenchmarkMatrix:
				m, err := env.Matrix()
				if err != nil {
					t.Fatalf("decoding the matrix measurement: %v", err)
				}
				got = report.Matrix{Result: m}.CSV()
			default:
				s, err := env.Standard()
				if err != nil {
					t.Fatalf("decoding the standard measurement: %v", err)
				}
				got = report.Standard{Result: s}.CSV()
			}

			if got != want {
				firstDiff(t, got, want)
				t.Error("the renderer no longer reproduces the committed CSV")
			}
		})
	}
}

// firstDiff narrows a corpus failure to one line, because the first difference
// in a 34-column CSV is the one that names the regression.
func firstDiff(t *testing.T, got, want string) {
	t.Helper()
	gotLines, wantLines := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(gotLines) && i < len(wantLines); i++ {
		if gotLines[i] != wantLines[i] {
			t.Errorf("%s: line %d\n got: %q\nwant: %q", corpusLine(i), i+1, gotLines[i], wantLines[i])
			return
		}
	}
	if len(gotLines) != len(wantLines) {
		at := len(gotLines)
		if len(wantLines) < at {
			at = len(wantLines)
		}
		t.Errorf("%s: rendered %d lines, the committed CSV has %d (first extra or missing line is %d)",
			corpusLine(at), len(gotLines), len(wantLines), at+1)
	}
}

func corpusName(t *testing.T, path string) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", "..", "benchmarks"))
	if err != nil {
		return filepath.Base(path)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.Base(path)
	}
	return strings.ReplaceAll(filepath.ToSlash(rel), "/", "_")
}

func corpusLine(i int) string {
	return "first difference"
}
