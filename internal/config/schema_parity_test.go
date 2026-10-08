package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// The config file has three copies of its key set and all three must say the
// same thing: the yaml struct tags in configfile.go (what the parser reads),
// allowedKeys (what checkKeys enforces with a friendly error), and
// schema/easysftp.schema.json (what an editor blesses through the modeline
// in docs/easysftp.example.yml). The schema sets additionalProperties: false
// throughout, so a key present in one copy but absent from another is drift
// in the direction that is hardest to notice: added to Go but not the schema,
// every editor marks a valid config invalid; removed from Go but left in the
// schema, an editor blesses a config the action then rejects at run time.
// Neither failure showed up anywhere before these tests (issue #233).

// goSections reflects over the yaml config structs and collects, per section
// path ("" is the file top level, "deployments.*" any named deployment),
// the sorted set of yaml keys that section struct declares.
func goSections() map[string][]string {
	sections := map[string][]string{}
	var walk func(typ reflect.Type, path string)
	walk = func(typ reflect.Type, path string) {
		keys := []string{}
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			name := strings.Split(f.Tag.Get("yaml"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			keys = append(keys, name)
			ft := f.Type
			if ft == reflect.TypeOf(yaml.Node{}) {
				// Deployments stays a raw node so the file ordering survives
				// (a Go map would shuffle it); the inner shape lives in
				// yamlDeployment.
				walk(reflect.TypeOf(yamlDeployment{}), name+".*")
				continue
			}
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() != reflect.Struct || taggedFields(ft) == 0 {
				// A struct with no yaml-tagged fields (autoInt) is a leaf
				// value, not a section to recurse into.
				continue
			}
			walk(ft, join(path, name))
		}
		slices.Sort(keys)
		sections[path] = slices.Compact(keys)
	}
	walk(reflect.TypeOf(yamlConfig{}), "")
	return sections
}

// taggedFields counts the yaml-tagged fields of a struct type.
func taggedFields(typ reflect.Type) int {
	n := 0
	for i := 0; i < typ.NumField(); i++ {
		if name := strings.Split(typ.Field(i).Tag.Get("yaml"), ",")[0]; name != "" && name != "-" {
			n++
		}
	}
	return n
}

// schemaDoc loads schema/easysftp.schema.json as a plain JSON document.
func schemaDoc(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "schema", "easysftp.schema.json"))
	if err != nil {
		t.Fatalf("read schema/easysftp.schema.json: %v", err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatalf("schema/easysftp.schema.json is not valid JSON: %v", err)
	}
	return root
}

// schemaSections collects the property names of every schema object that has
// some, keyed by section path like goSections. An additionalProperties
// schema (deployments) contributes the ".*" section.
func schemaSections(t *testing.T) map[string][]string {
	t.Helper()
	sections := map[string][]string{}
	var walk func(node any, path string)
	walk = func(node any, path string) {
		obj, ok := node.(map[string]any)
		if !ok {
			return
		}
		if props, ok := obj["properties"].(map[string]any); ok && len(props) > 0 {
			keys := []string{}
			for name, sub := range props {
				keys = append(keys, name)
				walk(sub, join(path, name))
			}
			slices.Sort(keys)
			sections[path] = keys
		}
		if ap, ok := obj["additionalProperties"].(map[string]any); ok {
			walk(ap, path+".*")
		}
	}
	walk(schemaDoc(t), "")
	return sections
}

func join(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

// diffKeys returns the keys of b missing from a and the keys of a missing
// from b, or nil, nil when the two sets agree.
func diffKeys(a, b []string) (missingFromA, missingFromB []string) {
	for _, k := range b {
		if !slices.Contains(a, k) {
			missingFromA = append(missingFromA, k)
		}
	}
	for _, k := range a {
		if !slices.Contains(b, k) {
			missingFromB = append(missingFromB, k)
		}
	}
	return missingFromA, missingFromB
}

// TestSchemaKeysMatchParser holds the config-file yaml struct tags and
// schema/easysftp.schema.json to the same key set, section by section, and
// reports the symmetric difference when they drift.
func TestSchemaKeysMatchParser(t *testing.T) {
	goKeys := goSections()
	schemaKeys := schemaSections(t)
	for path, want := range goKeys {
		got, ok := schemaKeys[path]
		if !ok {
			t.Errorf("config section %q is unknown to schema/easysftp.schema.json", path)
			continue
		}
		if missingSchema, missingParser := diffKeys(got, want); missingSchema != nil || missingParser != nil {
			t.Errorf("config section %q drifted from schema/easysftp.schema.json:\n  missing from schema: %v\n  missing from the parser: %v", path, missingSchema, missingParser)
		}
	}
	for path := range schemaKeys {
		if _, ok := goKeys[path]; !ok {
			t.Errorf("schema/easysftp.schema.json has config section %q, but the parser has none", path)
		}
	}
}

// TestAllowedKeysMatchParser holds allowedKeys (the runtime unknown-key
// check with the "did you mean" suggestions) to the same key set as the yaml
// struct tags, so an editor and the action can never disagree about what is
// a typo.
func TestAllowedKeysMatchParser(t *testing.T) {
	goKeys := goSections()
	for path, want := range allowedKeys {
		got, ok := goKeys[path]
		if !ok {
			t.Errorf("allowedKeys has section %q, but the parser has no such config section", path)
			continue
		}
		if dead, missing := diffKeys(got, want); dead != nil || missing != nil {
			t.Errorf("allowedKeys section %q drifted from the parser:\n  in allowedKeys but not in the parser: %v\n  missing from allowedKeys: %v", path, dead, missing)
		}
	}
	for path := range goKeys {
		if _, ok := allowedKeys[path]; !ok {
			t.Errorf("parser config section %q has no allowedKeys entry, so checkKeys cannot catch typos in it", path)
		}
	}
}

// TestSchemaEnumsMatchParser pins the schema's fixed values to what the
// parser accepts, so the two cannot silently diverge on a rename.
func TestSchemaEnumsMatchParser(t *testing.T) {
	root := schemaDoc(t)

	// version: the schema pins const 3 and applyYAML rejects anything else.
	props, _ := root["properties"].(map[string]any)
	version, _ := props["version"].(map[string]any)
	if version["const"] != any(float64(3)) {
		t.Errorf("schema properties.version const = %v, want 3", version["const"])
	}

	// mode: every enum value must be a valid Strategy, and every Strategy
	// must be in the enum.
	defs, _ := root["$defs"].(map[string]any)
	mode, _ := defs["mode"].(map[string]any)
	enum, _ := mode["enum"].([]any)
	if len(enum) == 0 {
		t.Fatal("schema $defs.mode has no enum")
	}
	for _, v := range enum {
		s, _ := v.(string)
		if !Strategy(s).valid() {
			t.Errorf("schema allows mode %q, but the parser rejects it", s)
		}
	}
	for _, s := range []string{string(StrategyOverlay), string(StrategySync), string(StrategyClean)} {
		if !slices.Contains(enum, any(s)) {
			t.Errorf("the parser accepts mode %q, but the schema enum does not list it", s)
		}
	}
}

// TestExampleConfigFileParses makes docs/easysftp.example.yml, the file
// users are told to copy, a tested artifact: it must load through the real
// parser, and the deployments it advertises must survive the load.
func TestExampleConfigFileParses(t *testing.T) {
	cfg := &Config{
		Concurrency:            defaultConcurrency,
		SftpRequestConcurrency: defaultRequestConcurrency,
		Retries:                defaultRetries,
		ManifestName:           DefaultManifestName,
	}
	if err := loadConfigFile(cfg, filepath.Join("..", "..", "docs", "easysftp.example.yml")); err != nil {
		t.Fatalf("docs/easysftp.example.yml no longer parses: %v", err)
	}
	// staging joins the example with PR #290: it is the merge-key
	// deployment (`<<: *website`) the file now advertises.
	want := []string{"website", "staging", "documentation", "robots"}
	got := make([]string, len(cfg.Uploads))
	for i, u := range cfg.Uploads {
		got[i] = u.Name
	}
	if !slices.Equal(got, want) {
		t.Errorf("the example's deployments changed: got %v, want %v (update this test alongside a deliberate example edit)", got, want)
	}
}
