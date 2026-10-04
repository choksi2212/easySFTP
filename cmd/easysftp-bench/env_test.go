package main

import (
	"reflect"
	"testing"
)

// The measuring commands are configured through the environment, and a
// misread variable produces a result set that looks valid while measuring the
// wrong thing (issue #239). These tests pin the parsing rules the axis
// display strings depend on; the command's main stays untested on purpose.

func TestEnvOrFallsBackOnEmpty(t *testing.T) {
	t.Setenv("ENV_TEST_LABEL", "")
	if got := envOr("ENV_TEST_LABEL", "fallback"); got != "fallback" {
		t.Fatalf("empty variable must fall back, got %q", got)
	}
	t.Setenv("ENV_TEST_LABEL", "value")
	if got := envOr("ENV_TEST_LABEL", "fallback"); got != "value" {
		t.Fatalf("set variable must win, got %q", got)
	}
}

func TestEnvPositive(t *testing.T) {
	t.Setenv("ENV_TEST_POS", "5")
	if v, err := envPositive("ENV_TEST_POS", 3); err != nil || v != 5 {
		t.Fatalf("set: got (%d, %v)", v, err)
	}
	t.Setenv("ENV_TEST_POS", "")
	if v, err := envPositive("ENV_TEST_POS", 3); err != nil || v != 3 {
		t.Fatalf("unset/empty: got (%d, %v)", v, err)
	}
	for _, bad := range []string{"0", "-1", "x", "1.5", " 7"} {
		t.Setenv("ENV_TEST_POS", bad)
		if _, err := envPositive("ENV_TEST_POS", 3); err == nil {
			t.Fatalf("value %q must be rejected", bad)
		}
	}
}

func TestEnvAxisParsesList(t *testing.T) {
	t.Setenv("ENV_TEST_AXIS", "1 2 4")
	values, display, err := envAxis("ENV_TEST_AXIS", "8 9")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(values, []int{1, 2, 4}) || display != "1 2 4" {
		t.Fatalf("got (%v, %q)", values, display)
	}
	t.Setenv("ENV_TEST_AXIS", "")
	if values, display, _ = envAxis("ENV_TEST_AXIS", "8 9"); !reflect.DeepEqual(values, []int{8, 9}) || display != "8 9" {
		t.Fatalf("fallback: got (%v, %q)", values, display)
	}
	t.Setenv("ENV_TEST_AXIS", "1 two")
	if _, _, err = envAxis("ENV_TEST_AXIS", "8"); err == nil {
		t.Fatal("non-integer axis value must be rejected")
	}
}

func TestEnvRequestAxisDisplayIsTheSweep(t *testing.T) {
	// Unset: the display string must name the fallback sweep that
	// actually ran, not the empty string the variable held. A stored
	// settings table that claims "easySFTP default" while three passes
	// ran is a document that lies about its own inputs.
	t.Setenv("ENV_TEST_RAXIS", "")
	values, display, err := envRequestAxis("ENV_TEST_RAXIS", "1 16 64")
	if err != nil {
		t.Fatal(err)
	}
	want := []*int{ptrInt(1), ptrInt(16), ptrInt(64)}
	if !reflect.DeepEqual(values, want) {
		t.Fatalf("values: got %v", values)
	}
	if display != "1 16 64" {
		t.Fatalf("display must echo the swept fallback, got %q", display)
	}

	// Whitespace-heavy set value: the display is the normalised field
	// list, so a table cell never carries stray tabs.
	t.Setenv("ENV_TEST_RAXIS", "  2   4 ")
	values, display, _ = envRequestAxis("ENV_TEST_RAXIS", "1 16 64")
	if !reflect.DeepEqual(values, []*int{ptrInt(2), ptrInt(4)}) {
		t.Fatalf("normalised values: got %v", values)
	}
	if display != "2 4" {
		t.Fatalf("display must be the joined fields, got %q", display)
	}

	// The literal token: a nil coordinate, kept in order.
	t.Setenv("ENV_TEST_RAXIS", "default 4")
	values, _, _ = envRequestAxis("ENV_TEST_RAXIS", "1 16 64")
	if len(values) != 2 || values[0] != nil || *values[1] != 4 {
		t.Fatalf("default token: got %v", values)
	}

	// Unparseable token: refused, not silently dropped.
	t.Setenv("ENV_TEST_RAXIS", "4 x")
	if _, _, err = envRequestAxis("ENV_TEST_RAXIS", "1"); err == nil {
		t.Fatal("unparseable request-axis value must be rejected")
	}
}

func ptrInt(v int) *int { return &v }
