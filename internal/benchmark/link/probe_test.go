package link

import (
	"encoding/json"
	"strings"
	"testing"
)

// The probe document is stored verbatim in results.json, which is an artifact
// and is committed: wrap decides what that permanent document says, so its
// edge cases are load-bearing (issue #239). summarize is the one line the job
// log gets to explain a probe with.

func TestWrapPrependsCoordinatesWithoutTouchingTheDocument(t *testing.T) {
	// A number the schema's own float64 would rewrite ("1.0" -> "1") and a
	// field no Go struct models must survive into the stored document.
	doc := []byte(`{"rtt_ms":{"p50":1.0},"future_field":{"deep":[1,2]}}`)
	wrapped, err := wrap("baseline", "start", doc)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(wrapped, &probe); err != nil {
		t.Fatalf("wrapped document is not JSON: %v", err)
	}
	if string(probe["rtt_ms"]) != `{"p50":1.0}` {
		t.Errorf("the measured numbers were re-encoded: %s", probe["rtt_ms"])
	}
	if _, ok := probe["future_field"]; !ok {
		t.Error("a field no Go struct models was dropped from the stored document")
	}
	if string(probe["profile"]) != `"baseline"` || string(probe["at"]) != `"start"` {
		t.Errorf("coordinates missing from the wrapped document: %s", wrapped)
	}
}

func TestWrapRejectsNonObjectDocuments(t *testing.T) {
	for name, doc := range map[string][]byte{
		"array":  []byte(`[1,2]`),
		"number": []byte(`42`),
		"empty":  []byte(``),
	} {
		if _, err := wrap("baseline", "start", doc); err == nil {
			t.Errorf("%s document was accepted", name)
		}
	}
	// The one object that is nothing but braces must not grow a dangling
	// comma; the wrapped result is valid JSON carrying exactly the two
	// coordinates and nothing else.
	wrapped, err := wrap("baseline", "start", []byte(`{}`))
	if err != nil {
		t.Fatalf("empty object: %v", err)
	}
	var only map[string]json.RawMessage
	if err := json.Unmarshal(wrapped, &only); err != nil {
		t.Fatalf("empty object wrapped as %s: %v", wrapped, err)
	}
	if len(only) != 2 {
		t.Errorf("empty object wrapped as %s, want exactly the two coordinates", wrapped)
	}
}

func TestWrapRejectsInvalidJSON(t *testing.T) {
	if _, err := wrap("baseline", "start", []byte(`{"p50":`)); err == nil {
		t.Error("truncated JSON was accepted")
	}
}

func TestSummarizeNamesTheProbeAndItsMeasurements(t *testing.T) {
	wrapped, err := wrap("+50ms/5mbit", "end", []byte(`{`+
		`"rtt_ms":{"p50":13.08,"samples":21},`+
		`"control":{"streams":4,"single_stream_mib_per_s":0.39,"n_stream_mib_per_s":1.04}}`))
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	got := summarize(wrapped)
	for _, want := range []string{
		"link probe +50ms/5mbit (end)",
		"rtt p50 13.08 ms",
		"control 0.39 MiB/s single / 1.04 MiB/s multi",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("summary %q does not contain %q", got, want)
		}
	}
}

func TestSummarizeHandlesAnEmptyDocument(t *testing.T) {
	// A probe that measured nothing is honest, an invented entry is not; the
	// summary must say n/a rather than 0 for what was never measured.
	wrapped, err := wrap("baseline", "start", []byte(`{}`))
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}
	got := summarize(wrapped)
	if !strings.Contains(got, "rtt p50 n/a ms") || !strings.Contains(got, "control n/a MiB/s single / n/a MiB/s multi") {
		t.Errorf("empty document summarized as %q", got)
	}
}

func TestSummarizeReportsUnreadableDocuments(t *testing.T) {
	if got := summarize([]byte(`not json`)); got != "link probe: unreadable" {
		t.Errorf("unreadable document summarized as %q", got)
	}
}
