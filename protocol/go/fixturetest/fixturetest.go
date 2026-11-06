// Package fixturetest runs the cases of protocol/fixtures/ against a rule.
package fixturetest


import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)


// --- GLOBALS ---

// referenceKeys are the expect keys whose values name events or authors.
var referenceKeys = []string{
	"head", "chain", "losingBranch", "withdrawnChain", "responsesOnHead",
	"respondedTo", "countedVote", "stored", "counted", "labelledProposals",
	"retained",
}


// --- CODE ---

// File is one fixture file.
type File struct {
	Topic string `json:"topic"`
	Rule  string `json:"rule"`
	Cases []Case `json:"cases"`
}


// Case is one behavioural case of a file.
type Case struct {
	Name       string                     `json:"name"`
	Why        string                     `json:"why"`
	Since      string                     `json:"since"`
	Identities map[string]json.RawMessage `json:"identities"`
	Arrival    []Arrival                  `json:"arrival"`
	Expect     map[string]any             `json:"expect"`
}


// Arrival is one event, in the order the node observed it.
type Arrival struct {
	ID       string          `json:"id"`
	Author   string          `json:"author"`
	SignedBy string          `json:"signedBy"`
	Type     string          `json:"type"`
	At       string          `json:"at"`
	Payload  json.RawMessage `json:"payload"`
}


// Load reads and checks protocol/fixtures/<name>.json from the repository root.
func Load(t *testing.T, name string) File {
	t.Helper()

	file, err := load("../../fixtures", name)
	if err != nil {
		t.Fatalf("fixture %s: %v", name, err)
	}

	return file
}


// load reads and checks <dir>/<name>.json.
func load(dir, name string) (File, error) {
	raw, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return File{}, err
	}

	return parse(dir, raw)
}


// Run runs rule once per case, each as a subtest named after the case.
func Run(t *testing.T, file File, rule func(t *testing.T, c Case)) {
	t.Helper()

	if rule == nil {
		t.Skipf("%s: no rule implements %s yet", file.Topic, file.Rule)
	}

	for _, c := range file.Cases {
		t.Run(c.Name, func(t *testing.T) {
			if c.Since != "" {
				t.Skipf("runs since %s, not in this version", c.Since)
			}

			rule(t, c)
		})
	}
}


// parse validates raw against the fixture schema and decodes it.
func parse(dir string, raw []byte) (File, error) {
	if err := checkSchema(dir, raw); err != nil {
		return File{}, err
	}

	var file File
	if err := json.Unmarshal(raw, &file); err != nil {
		return File{}, err
	}

	if err := checkCases(file); err != nil {
		return File{}, err
	}

	return file, nil
}


// checkCases applies the checks beyond the schema to every case of file.
func checkCases(file File) error {
	for _, c := range file.Cases {
		known, err := checkArrival(c)
		if err != nil {
			return err
		}

		if err := checkReferences(c, known); err != nil {
			return err
		}

		if err := checkWeights(c); err != nil {
			return err
		}
	}

	return nil
}


// checkArrival refuses a repeated id and returns the ids and authors of c.
func checkArrival(c Case) (map[string]bool, error) {
	known := map[string]bool{}

	for _, a := range c.Arrival {
		if known[a.ID] {
			return nil, fmt.Errorf("case %q: id %q repeats", c.Name, a.ID)
		}

		known[a.ID] = true
	}

	// a rotation chain in expect names authors and their successors
	for _, a := range c.Arrival {
		known[a.Author] = true

		if next := successor(a); next != "" {
			known[next] = true
		}
	}

	return known, nil
}


// successor returns the successor key a rotation names, or "" for no one.
func successor(a Arrival) string {
	var payload struct {
		Successor string `json:"successor"`
	}

	// a payload that is not an object names no successor
	_ = json.Unmarshal(a.Payload, &payload)

	return payload.Successor
}


// checkReferences refuses a name in expect that arrival never introduces.
func checkReferences(c Case, known map[string]bool) error {
	for _, ref := range references(c.Expect) {
		if !known[ref] {
			return fmt.Errorf("case %q: expect names %q, absent from arrival",
			                  c.Name, ref)
		}
	}

	return nil
}


// references lists the names expect gives to events and authors.
func references(expect map[string]any) []string {
	var refs []string

	for _, key := range referenceKeys {
		refs = append(refs, names(expect[key])...)
	}

	for name := range asMap(expect["historical"]) {
		refs = append(refs, name)
	}

	for _, label := range asList(expect["labelsOnHead"]) {
		refs = append(refs, names(asMap(label)["id"])...)
	}

	return refs
}


// names returns the strings of v, which is one string or a list of them.
func names(v any) []string {
	if s, ok := v.(string); ok {
		return []string{s}
	}

	var out []string

	for _, item := range asList(v) {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}

	return out
}


// asList returns v as a list, or nil when it is not one.
func asList(v any) []any {
	list, _ := v.([]any)

	return list
}


// asMap returns v as an object, or nil when it is not one.
func asMap(v any) map[string]any {
	object, _ := v.(map[string]any)

	return object
}


// checkWeights refuses a proposal whose author has no weight declared.
func checkWeights(c Case) error {
	if len(c.Identities) == 0 {
		return nil
	}

	for _, a := range c.Arrival {
		_, declared := c.Identities[a.Author]
		if a.Type == "entity.updated" && !declared {
			return fmt.Errorf("case %q: %q proposes with no weight",
			                  c.Name, a.Author)
		}
	}

	return nil
}


// checkSchema validates raw against protocol/fixtures/fixture.schema.json.
func checkSchema(dir string, raw []byte) error {
	compiler := jsonschema.NewCompiler()

	schema, err := compiler.Compile(filepath.Join(dir, "fixture.schema.json"))
	if err != nil {
		return err
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return err
	}

	return schema.Validate(doc)
}
