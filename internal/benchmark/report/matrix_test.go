package report_test

import (
	"strings"
	"testing"

	"github.com/eiserv/easySFTP/internal/benchmark/report"
	"github.com/eiserv/easySFTP/internal/benchmark/schema"
)

// matrixFixture is a two-scenario, one-profile sweep with both builds, an auto
// run, a canary triple and a delete sweep, small enough to read but carrying
// every block the summary renders.
func matrixFixture() report.Matrix {
	rc := 16
	result := &schema.Matrix{
		CandidateRef:   "cand (abcdef1)",
		BaselineRef:    "base (1234567)",
		ReferenceLabel: "baseline",
		Repeats:        2,
		Runner:         "self-hosted, Linux 6.1.0, 4 cpu",
		Link:           linkFixture(),
		Axes: schema.Axes{
			LinkProfiles:       []string{"baseline"},
			Connections:        []int{1, 2},
			Concurrency:        []int{1, 4},
			RequestConcurrency: []*int{nil, &rc},
		},
		Cells: []schema.Cell{
			{
				Scenario: "bulk", Label: "candidate", Ref: "cand (abcdef1)", LinkProfile: "baseline",
				Connections: 1, Concurrency: 1, Repeats: 2, Files: 2000, Bytes: 8192000,
				MedianMS: 246659, MinMS: 246659, MaxMS: 261103, MadMS: ptr(0), MiBPerS: 0.03, FilesPerS: 8.11,
				RequestConcurrencyUsed: ptr(16),
			},
			{
				Scenario: "bulk", Label: "candidate", Ref: "cand (abcdef1)", LinkProfile: "baseline",
				Connections: 2, Concurrency: 4, Repeats: 2, Files: 2000, Bytes: 8192000,
				MedianMS: 50000, MinMS: 50000, MaxMS: 52000, MadMS: ptr(100), MiBPerS: 0.16, FilesPerS: 40,
				NetWriteBytes: ptr(9605940), ConnectionsRefused: ptr(1),
				RequestConcurrencyUsed: ptr(16),
			},
			{
				Scenario: "bulk", Label: "baseline", Ref: "base (1234567)", LinkProfile: "baseline",
				Connections: 1, Concurrency: 1, Repeats: 2, Files: 2000, Bytes: 8192000,
				MedianMS: 260000, MinMS: 260000, MaxMS: 270000, MadMS: nil, MiBPerS: 0.03, FilesPerS: 7.69,
				ConnectionsRefused: ptr(1),
			},
		},
		Comparison: []schema.MatrixCompare{
			{
				Scenario: "bulk", Label: "candidate", ReferenceLabel: "baseline", LinkProfile: "baseline",
				Connections: 1, Concurrency: 1, MedianMS: 246659, ReferenceMedianMS: 260000,
				DeltaMS: -13341, DeltaPercent: ptr(-5.13),
			},
			{
				Scenario: "bulk", Label: "candidate", ReferenceLabel: "baseline", LinkProfile: "baseline",
				Connections: 1, Concurrency: 1, MedianMS: 246659, ReferenceMedianMS: 246659,
				DeltaMS: 0, DeltaPercent: ptr(0),
			},
			{
				// The best cell of the group sits at a different coordinate from
				// the worst, so the deduplication in the comparison table keeps
				// both rows: one improvement and one regression.
				Scenario: "bulk", Label: "candidate", ReferenceLabel: "baseline", LinkProfile: "baseline",
				Connections: 2, Concurrency: 4, MedianMS: 50000, ReferenceMedianMS: 233000,
				DeltaMS: -13341, DeltaPercent: ptr(5.86),
			},
			{
				// A scenario with a single measured cell: its only entry is
				// both the worst and the best of its group, which is where
				// the table's deduplication earns its keep.
				Scenario: "single", Label: "candidate", ReferenceLabel: "baseline", LinkProfile: "baseline",
				Connections: 1, Concurrency: 4, MedianMS: 400, ReferenceMedianMS: 410,
				DeltaMS: -10, DeltaPercent: ptr(-2.44),
			},
		},
		Scaling: []schema.Scaling{
			{
				Scenario: "bulk", Label: "candidate", LinkProfile: "baseline",
				Points: []schema.Point{
					{Connections: 1, Concurrency: 1, MedianMS: 246659, MiBPerS: 0.03},
					{Connections: 2, Concurrency: 4, MedianMS: 50000, MiBPerS: 0.16},
				},
				Best: schema.Best{Connections: 2, Concurrency: 4, MedianMS: 50000, MiBPerS: 0.16, FilesPerS: 40},
			},
			{
				// The best cell sits on the largest swept concurrency, which
				// must print the cut-off warning rather than stay quiet.
				Scenario: "single", Label: "candidate", LinkProfile: "baseline",
				Best:          schema.Best{Connections: 1, Concurrency: 4, MedianMS: 400, MiBPerS: 1.2, FilesPerS: 2.5},
				BestAtAxisMax: []string{"concurrency"},
			},
		},
		Auto: []schema.Auto{
			{
				Scenario: "bulk", Label: "candidate", Ref: "cand (abcdef1)", LinkProfile: "baseline",
				Repeats: 2, MedianMS: 51000, MinMS: 51000, MaxMS: 52000, Files: 2000, Bytes: 8192000,
				MiBPerS: 0.15, FilesPerS: 39.2,
				Chosen: schema.Chosen{
					Connections: ptr(2), Concurrency: ptr(4), RequestConcurrency: ptr(16),
					InitialConnections: ptr(1), InitialConcurrency: ptr(1), InitialRequestConcurrency: ptr(4),
					Changes: ptr(3), SpreadIncreases: ptr(2), SpreadDecreases: ptr(1),
				},
				ConnectionsOpened: ptr(1), ConnectionsUsed: ptr(1), ConnectionsRefused: ptr(1),
				Workload: &schema.AutoWorkload{
					Files: ptr(2000), Bytes: ptr(8192000), LargestBytes: ptr(8192), Probes: ptr(40),
					P50Bytes: ptr(4096), P90Bytes: ptr(4096), SmallFiles: ptr(1990),
					RTTMS: ptr(13.2), HandshakeMS: ptr(400.1), BDPBytes: ptr(163840),
				},
				Best:               &schema.Best{Connections: 2, Concurrency: 4, RequestConcurrency: &rc, MedianMS: 50000, MiBPerS: 0.16, FilesPerS: 40},
				ChosenInGrid:       true,
				ChosenCellMedianMS: ptr(50000),
				RegretPercent:      ptr(2),
			},
		},
		Canary: []schema.Canary{
			{LinkProfile: "baseline", At: "start", Scenario: "bulk", Connections: 1, Concurrency: 1, DurationMS: 240000, Files: 2000, Bytes: 8192000},
			{LinkProfile: "baseline", At: "mid", Scenario: "bulk", Connections: 1, Concurrency: 1, DurationMS: 264000, Files: 2000, Bytes: 8192000},
			{LinkProfile: "baseline", At: "end", Scenario: "bulk", Connections: 1, Concurrency: 1, DurationMS: 252000, Files: 2000, Bytes: 8192000},
		},
		Deletes: []schema.MatrixDelete{
			{
				Scenario: "bulk", Label: "candidate", LinkProfile: "baseline", Connections: 2, Concurrency: 4,
				DeleteStats: schema.DeleteStats{
					Sweeps: 1, FilesDeleted: 2000, MedianMS: 5000, MinMS: 5000, MaxMS: 5000, DeletesPerS: 400,
					Phases: []schema.Phase{{Name: "remote_scan", MedianMS: 50}, {Name: "delete_sweep", MedianMS: 4950}},
					Operations: []schema.Operation{
						{Name: "sftp_remove", Count: ptr(2000), MedianTotalMS: ptr(4000), P50MS: ptr(2)},
						{Name: "sftp_rmdir", Count: ptr(100), MedianTotalMS: ptr(950), P50MS: ptr(9)},
					},
				},
			},
		},
	}
	return report.Matrix{
		Result:             result,
		CandidateRef:       "cand (abcdef1)",
		BaselineRef:        "base (1234567)",
		Repeats:            2,
		RunnerDisplay:      "Linux 6.1.0, 4 cpu",
		ConnectionsDisplay: "1 2",
		ConcurrencyDisplay: "1 4",
		RequestDisplay:     "default 16",
		Labels:             []string{"candidate", "baseline"},
		LinkProfiles:       []string{"baseline"},
		Scenarios: []report.MatrixScenario{
			{
				Name: "bulk", Mode: "overlay", Description: "2000 4 KiB files",
				ConnectionsDisplay: "1 2", ConcurrencyDisplay: "1 4", RequestDisplay: "default 16",
				Connections: []int{1, 2}, Concurrency: []int{1, 4},
				RequestConcurrency: []*int{nil, &rc},
			},
			{
				Name: "single", Mode: "overlay", Description: "one 32 MiB file",
				ConnectionsDisplay: "1", ConcurrencyDisplay: "1 4", RequestDisplay: "16",
				Connections: []int{1}, Concurrency: []int{1, 4},
				RequestConcurrency: []*int{&rc},
			},
		},
		Canary:      schema.Canary{Scenario: "bulk", Connections: 1, Concurrency: 1},
		HasBaseline: true,
	}
}

func TestMatrixCSVHeaderAndColumnOrder(t *testing.T) {
	got := matrixFixture().CSV()
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) != 4 { // header + three cells
		t.Fatalf("rendered %d lines, want 4 (header + 3 cells)", len(lines))
	}
	wantHeader := `"scenario","build","ref","link_profile","rtt_p50_ms","control_single_mib_per_s","connections","concurrency","request_concurrency","request_concurrency_used","repeats","files","bytes","median_ms","min_ms","max_ms","mad_ms","mib_per_s","files_per_s","upload_phase_ms","net_write_bytes","user_cpu_ms","sys_cpu_ms","cpu_percent","max_rss_bytes","go_gc_count","go_peak_goroutines","connections_opened","connections_used","connections_refused","reconnects","retries","errors","failed_runs"`
	if lines[0] != wantHeader {
		t.Errorf("header\n got: %s\nwant: %s", lines[0], wantHeader)
	}
	// Cells follow the JSON's order, one flat row per cell.
	if !strings.HasPrefix(lines[1], `"bulk","candidate","cand (abcdef1)","baseline",13.08,0.39,1,1,`) {
		t.Errorf("cell 1 does not start with its coordinates and probe: %s", lines[1])
	}
	// A null request_concurrency is the pass that set nothing; it renders as
	// an empty field, while request_concurrency_used carries the value the
	// run actually used.
	fields := strings.Split(lines[1], ",")
	if len(fields) != 34 {
		t.Fatalf("cell 1 has %d fields, want 34", len(fields))
	}
	if fields[8] != "" || fields[9] != "16" {
		t.Errorf("request_concurrency rendered %q and used %q, want empty and 16", fields[8], fields[9])
	}
}

func TestMatrixMarkdownPinsTheDisplayRules(t *testing.T) {
	md := matrixFixture().Markdown()

	// The settings table prints every axis as the script was given it.
	for _, want := range []string{
		"| connections | 1 2 |",
		"| concurrency | 1 4 |",
		"| request_concurrency | default 16 |",
		"| Scenarios | bulk single |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("the settings table lost %q", want)
		}
	}
	// The scenario table prints each scenario's own axis values.
	if !strings.Contains(md, "| `single` | overlay | one 32 MiB file | 1 | 1 4 | 16 |") {
		t.Error("the scenario table did not print single's own axes")
	}

	// The grid prints the request_concurrency pass in its title and one row
	// per connections value, in axis order.
	if !strings.Contains(md, "#### `bulk` / candidate / baseline / request_concurrency 16") {
		t.Error("the shaped-pass grid title went missing")
	}
	if !strings.Contains(md, "| connections \\ concurrency | 1 | 4 |") {
		t.Error("the grid header did not print the concurrency axis")
	}
	// A cell that exists prints median, throughput and a refused marker; a
	// cell that does not prints "-".
	if !strings.Contains(md, "50000 ms<br>0.16 MiB/s<br>1 refused") {
		t.Error("a refused cell did not print its refused count")
	}

	// The best-cell table and the cut-off warning.
	if !strings.Contains(md, "| bulk | candidate | baseline | 2 | 4 | default | 50000 ms | 0.16 | 40 | no |") {
		t.Error("the best-cell row did not render with request_concurrency as default")
	}
	if !strings.Contains(md, "**The optimum sits on the edge of the grid** for single/candidate/baseline: concurrency.") {
		t.Error("a best cell on the axis maximum did not print the cut-off warning")
	}

	// Candidate against baseline: worst and best per (profile, scenario),
	// deduplicated and sorted by coordinates, with the delta's sign.
	if !strings.Contains(md, "| bulk | baseline | 1 | 1 | 246659 ms | 260000 ms | -5.13% |") {
		t.Error("the worst cell did not render with its negative delta")
	}
	if !strings.Contains(md, "| bulk | baseline | 2 | 4 | 50000 ms | 233000 ms | +5.86% |") {
		t.Error("the best cell did not gain the explicit + on a regression")
	}
	// The worst and the best sit at different coordinates, so both rows
	// survive the table's deduplication; neither stands in for the other.
	if count := strings.Count(md, "ms | 233000 ms |"); count != 1 {
		t.Errorf("the best cell appears in %d comparison rows, want 1", count)
	}
	// A single-cell group contributes that cell as both its worst and its
	// best; the deduplication owes one row for it, not two.
	if count := strings.Count(md, "| single | baseline | 1 | 4 | 400 ms | 410 ms | -2.44% |"); count != 1 {
		t.Errorf("the single-cell group rendered %d rows, want 1", count)
	}

	// The auto-cost table: chosen settings, changes split by direction,
	// carried marked short and refused, regret, and the in-grid control.
	if !strings.Contains(md, "| bulk | baseline | 2/4/16 | 1/1/4 | 3 (+2/-1) | 1/1 (short) refused 1 | 51000 ms | 2/4/16 | 50000 ms | +2% | 50000 ms |") {
		t.Error("the auto-cost table lost its row")
	}
	if !strings.Contains(md, "3 (+2/-1)") {
		t.Error("changes did not render as total (+up/-down)")
	}
	if !strings.Contains(md, "1/1 (short) refused 1") {
		t.Error("carried did not render used/opened with its short and refused markers")
	}
	// The workload block prints the input side of the decision.
	if !strings.Contains(md, "<details><summary>What the policy was looking at</summary>") {
		t.Error("the workload details block went missing")
	}
	if !strings.Contains(md, "| bulk | baseline | 2000 | 8192000 | 4096 | 4096 | 8192 | 1990 | 40 | 13.2 ms | 400.1 ms | 163840 |") {
		t.Error("the workload row did not render its fields")
	}

	// The canary table computes the spread between best and worst.
	if !strings.Contains(md, "| baseline | 240000 ms | 264000 ms | 252000 ms | 10% |") {
		t.Error("the canary row did not render its three readings and spread")
	}

	// The delete sweeps render per-cell coordinates and the two phase
	// columns through the ms form, plus the p50 of both operations.
	if !strings.Contains(md, "| bulk | candidate | baseline | 2 | 4 | 2000 | 5000 ms | 400 | 50 ms | 4950 ms | 2 ms | 9 ms |") {
		t.Error("the delete-sweep row did not render its phases and p50s")
	}

	// The refused warning names the sweep total over the cells only: the
	// auto run's refusal is its own row's marker, not part of the sum.
	if !strings.Contains(md, "**2 connection(s) were refused by the server** across the sweep") {
		t.Error("the sweep's refused-connections warning went missing")
	}

	// An empty request_concurrency axis prints "easySFTP default" in the
	// settings table, not an empty cell.
	m := matrixFixture()
	m.RequestDisplay = ""
	if s := m.Markdown(); !strings.Contains(s, "| request_concurrency | easySFTP default |") {
		t.Error("an empty request display did not render as easySFTP default")
	}
}
