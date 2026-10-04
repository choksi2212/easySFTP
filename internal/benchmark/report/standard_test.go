package report_test

import (
	"strings"
	"testing"

	"github.com/eiserv/easySFTP/internal/benchmark/report"
	"github.com/eiserv/easySFTP/internal/benchmark/schema"
)

// The corpus test pins the renderer against real stored results; this fixture
// covers the branches the corpus happens not to contain: every null form, the
// refused-connections warning, and the scenario ordering the loop owns
// (issue #239). Each expected value states which rule it pins, so a failure
// names the behaviour that regressed.

func ptr(f float64) *float64 { return &f }

func boolPtr(b bool) *bool { return &b }

// standardFixture is one standard result with two scenarios and two builds,
// carrying a null in every nullable position the renderer reads.
func standardFixture() report.Standard {
	result := &schema.Standard{
		CandidateRef:   "cand (abcdef1)",
		BaselineRef:    "base (1234567)",
		Repeats:        3,
		Runner:         "self-hosted, Linux 6.1.0, 4 cpu",
		ReferenceLabel: "baseline",
		Scenarios:      map[string]string{"small": "300 small files", "single": "one 32 MiB file"},
		Link:           linkFixture(),
		Results: []schema.Result{
			{
				Label: "candidate", Ref: "cand (abcdef1)", Scenario: "small", LinkProfile: "baseline",
				Repeats: 3, Files: 300, Bytes: 1228800, MedianMS: 3633, MinMS: 3482, MaxMS: 3857,
				MadMS: ptr(151), MiBPerS: 0.32, FilesPerS: 82.58,
				Process: &schema.Process{
					UserCPUMS: ptr(201.09), SysCPUMS: ptr(232.23), CPUPercent: ptr(3.54),
					MaxRSSBytes: ptr(11681792), GoTotalAllocBytes: ptr(2097152),
					GoGCCount: ptr(2), GoGCPauseTotalMS: ptr(4.5), GoPeakGoroutines: ptr(109),
					NetWriteBytes: ptr(12280892),
				},
				Counters: schema.Counters{"connections_opened": ptr(2), "connections_refused": ptr(1)},
				Phases:   []schema.Phase{{Name: "upload", MedianMS: 3500}, {Name: "remote_scan", MedianMS: 12}},
				Operations: []schema.Operation{
					{Name: "sftp_write", Count: ptr(300), MedianTotalMS: ptr(2100.5), AvgMS: ptr(7), P50MS: ptr(7.1), P90MS: ptr(9), P99MS: ptr(12), MaxMS: ptr(40)},
					{Name: "sftp_stat", Count: ptr(0)}, // below one: never printed, not as zero
				},
				RefusedConnections: ptr(1),
			},
			{
				// A metrics file that never appeared: every resource column is
				// null and must render as the word "null" in the Markdown and
				// as an empty field in the CSV, never as 0 (see raw in
				// report.go for why the Markdown keeps the word).
				Label: "baseline", Ref: "base (1234567)", Scenario: "small", LinkProfile: "baseline",
				Repeats: 1, Files: 300, Bytes: 1228800, MedianMS: 4200, MinMS: 4200, MaxMS: 4200,
				MiBPerS: 0.28, FilesPerS: 71.43,
			},
			{
				Label: "candidate", Ref: "cand (abcdef1)", Scenario: "single", LinkProfile: "baseline",
				Repeats: 3, Files: 1, Bytes: 33554432, MedianMS: 48954, MinMS: 48705, MaxMS: 50432,
				MadMS: ptr(249), MiBPerS: 0.65, FilesPerS: 0.02,
				Process:  &schema.Process{UserCPUMS: ptr(411.97)},
				Counters: schema.Counters{"connections_opened": ptr(1)},
				Phases:   []schema.Phase{{Name: "upload", MedianMS: 48900}},
			},
		},
		Comparison: []schema.Comparison{
			{
				Scenario: "small", Label: "candidate", LinkProfile: "baseline", ReferenceLabel: "baseline",
				MedianMS: 3633, ReferenceMedianMS: 4200, DeltaMS: -567, DeltaPercent: ptr(-13.5),
				ReferenceMadMS: ptr(151), WithinNoise: boolPtr(false),
			},
			{
				Scenario: "single", Label: "candidate", LinkProfile: "baseline", ReferenceLabel: "baseline",
				MedianMS: 48954, ReferenceMedianMS: 48954, DeltaMS: 0, DeltaPercent: ptr(2.4),
			},
		},
		Deletes: []schema.StandardDelete{
			{
				Label: "candidate", Scenario: "small", LinkProfile: "baseline",
				DeleteStats: schema.DeleteStats{
					Sweeps: 3, FilesDeleted: 300, MedianMS: 950, MinMS: 900, MaxMS: 1000, MadMS: ptr(20),
					DeletesPerS: 315.79,
					Phases:      []schema.Phase{{Name: "remote_scan", MedianMS: 15}, {Name: "delete_sweep", MedianMS: 935}},
					Operations: []schema.Operation{
						{Name: "sftp_remove", Count: ptr(300), MedianTotalMS: ptr(900), P50MS: ptr(2.5), P90MS: ptr(3), P99MS: ptr(4), MaxMS: ptr(9)},
						{Name: "sftp_rmdir", Count: ptr(12)},
					},
				},
			},
		},
	}
	return report.Standard{
		Result:        result,
		CandidateRef:  "cand (abcdef1)",
		BaselineRef:   "base (1234567)",
		Repeats:       3,
		RunnerDisplay: "Linux 6.1.0, 4 cpu",
		Settings:      "easySFTP defaults",
		Labels:        []string{"candidate", "baseline"},
		Scenarios:     []string{"small", "single"},
		LinkProfiles:  []string{"baseline"},
	}
}

func TestStandardCSVHeaderAndColumnOrder(t *testing.T) {
	got := standardFixture().CSV()
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 4 { // header + three results
		t.Fatalf("rendered %d lines, want 4 (header + 3 rows)", len(lines))
	}
	wantHeader := `"scenario","build","ref","link_profile","rtt_p50_ms","control_single_mib_per_s","repeats","files","bytes","median_ms","min_ms","max_ms","mad_ms","mib_per_s","files_per_s","user_cpu_ms","sys_cpu_ms","cpu_percent","max_rss_bytes","go_gc_count","go_peak_goroutines","net_write_bytes","connections_opened","connections_refused","retries","errors","failed_runs"`
	if lines[0] != wantHeader {
		t.Errorf("header\n got: %s\nwant: %s", lines[0], wantHeader)
	}
	// Rows follow the JSON's Results order, not the loop order the Markdown
	// uses, because a spreadsheet reads the file and not the summary.
	if !strings.HasPrefix(lines[1], `"small","candidate","cand (abcdef1)","baseline",`) {
		t.Errorf("row 1 does not start with the candidate's small row: %s", lines[1])
	}
	// The baseline row's metrics never appeared, so its resource columns
	// must be empty fields rather than zeros.
	fields := strings.Split(lines[2], ",")
	if len(fields) != 27 {
		t.Fatalf("row 3 has %d fields, want 27", len(fields))
	}
	for _, i := range []int{15, 16, 17, 18, 19, 20, 21} { // user_cpu_ms .. net_write_bytes
		if fields[i] != "" {
			t.Errorf("a null metric rendered as %q in column %d, want an empty field", fields[i], i+1)
		}
	}
	// Counters render their real values, and the refused count is carried.
	if !strings.HasSuffix(lines[1], `,2,1,0,0,0`) {
		t.Errorf("candidate row tail %q, want opened 2, refused 1, retries/errors/failed 0", lines[1])
	}
	// The start probe belongs to the row's own profile, so the rtt and the
	// control throughput of that profile land on every row of the profile.
	if !strings.Contains(lines[1], `"baseline",13.08,0.39,`) {
		t.Errorf("the row did not pick up its profile's start probe: %s", lines[1])
	}
}

func TestStandardMarkdownPinsTheDisplayRules(t *testing.T) {
	md := standardFixture().Markdown()

	// The settings table: an empty BaselineRef prints "none" and an empty
	// LinkRequested prints "the real line".
	if !strings.Contains(md, "| Baseline | `base (1234567)` |") {
		t.Error("the settings table lost the baseline ref")
	}
	empty := report.Standard{Result: &schema.Standard{Scenarios: map[string]string{}}}
	if s := empty.Markdown(); !strings.Contains(s, "| Baseline | `none` |") || !strings.Contains(s, "| Link profiles | the real line |") {
		t.Error("an empty baseline/link did not render as none/the real line")
	}

	// The throughput table's column order is part of the committed document:
	// a header reshuffle would rewrite every future results.md at the next
	// store, so the exact header line is pinned.
	if !strings.Contains(md, "| Scenario | Build | Profile | Files | Size | Median | Min | Max | MAD | MiB/s | files/s | Retries | Errors | Failed runs | Delta |") {
		t.Error("the throughput table header changed")
	}
	// The throughput table loops scenario, then build, then profile, which is
	// the script's order rather than the aggregation's.
	throughput := section(md, "### Throughput")
	smallIdx := strings.Index(throughput, "| small | candidate |")
	singleIdx := strings.Index(throughput, "| single | candidate |")
	baselineIdx := strings.Index(throughput, "| small | baseline |")
	if smallIdx < 0 || singleIdx < 0 || baselineIdx < 0 {
		t.Fatal("a throughput row went missing")
	}
	if !(smallIdx < baselineIdx && baselineIdx < singleIdx) {
		t.Error("throughput rows are not in scenario-then-build order")
	}

	// MAD is a nullable duration: null renders "-", a value renders "n ms".
	if !strings.Contains(md, "| 151 ms |") {
		t.Error("a measured MAD did not render as \"151 ms\"")
	}
	if !strings.Contains(md, "| 4200 ms | 4200 ms | 4200 ms | - |") {
		t.Error("a null MAD did not render as \"-\"")
	}

	// Delta: negative is a plain number, positive gains a plus, and the
	// reference build's own row has no comparison entry, so it is "-".
	if !strings.Contains(md, "| -13.5% |") {
		t.Error("a negative delta lost its percent form")
	}
	if !strings.Contains(md, "| +2.4% |") {
		t.Error("a positive delta did not gain the explicit +")
	}

	// Resources: a metrics file that never appeared renders its time and
	// count columns as the word "null" (raw, which is what the stored
	// document contains) and its byte columns as "- MiB" (mibOf), which is
	// what the shell printed for a run without metrics.
	if !strings.Contains(md, "| small | baseline | baseline | null ms | null ms | null% | - MiB | - MiB | null | null ms | null | - MiB |") {
		t.Error("a missing metrics document did not render its nulls as the word null")
	}

	// The refused-connections warning appears once and names the count.
	if !strings.Contains(md, "**1 connection(s) were refused by the server**") {
		t.Error("the refused-connections warning went missing")
	}
	// And it stays silent when nothing was refused.
	silent := standardFixture()
	silent.Result.Results[0].RefusedConnections = nil
	silent.Result.Results[0].Counters = schema.Counters{"connections_opened": ptr(2)}
	if s := silent.Markdown(); strings.Contains(s, "were refused by the server") {
		t.Error("a run with no refused connections still warned")
	}

	// Phases and operations print only what was measured.
	if !strings.Contains(md, "| candidate | baseline | upload | 3500 ms |") {
		t.Error("a measured phase went missing from the details block")
	}
	if strings.Contains(md, "sftp_stat") {
		t.Error("a zero-count operation printed anyway")
	}
	if !strings.Contains(md, "| baseline | sftp_remove | 300 | 900 ms |") {
		t.Error("the delete-sweep operation block lost its row")
	}

	// MiB rendering rounds to one decimal, the way the shell's jq did.
	if !strings.Contains(md, "| 300 | 1.2 MiB |") {
		t.Error("a 1228800-byte payload did not render as 1.2 MiB")
	}

	// The delete sweep renders its two phase columns through the ms form.
	if !strings.Contains(md, "| 15 ms | 935 ms |") {
		t.Error("the delete sweep phases did not render with the ms suffix")
	}
}

// section cuts the block of a summary between two "### " headings, so one
// rule can be asserted without the whole document being one assertion.
func section(md, heading string) string {
	start := strings.Index(md, heading)
	if start < 0 {
		return ""
	}
	rest := md[start+len(heading):]
	end := strings.Index(rest, "\n### ")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func linkFixture() *schema.Link {
	reason := "tc is not installed"
	iface := "eth0"
	return &schema.Link{
		Iface: &iface,
		Shaping: schema.Shaping{
			Available: false, Reason: &reason,
			Requested: []string{"baseline"}, Applied: []string{"baseline"},
		},
		Probes: []schema.Probe{
			{
				Profile: "baseline", At: "start", HandshakeMS: ptr(403.66),
				RTTMS:    &schema.RTT{P50: ptr(13.08), P90: ptr(13.36), Min: ptr(12.89), Max: ptr(14.72), Samples: 21},
				Control:  &schema.Control{Streams: 4, Bytes: 8388608, SingleStreamMiBPerS: ptr(0.39), NStreamMiBPerS: ptr(1.04)},
				HostLoad: &schema.HostLoad{Available: true, Load1: ptr(1.29)},
			},
		},
	}
}
