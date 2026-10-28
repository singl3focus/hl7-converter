package hl7converter_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hl7 "github.com/singl3focus/hl7-converter/v2"
)

func TestCheckMappingsFixtures(t *testing.T) {
	results, err := hl7.CheckMappings(filepath.Join("examples", "mappings", "suite.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatal(results)
	}
	for _, result := range results {
		if !result.Passed {
			t.Fatalf("%+v", result)
		}
	}
}

func TestCheckMappingsFailuresAndStrictFormat(t *testing.T) {
	dir := t.TempDir()
	input, output := auditMods("|")
	output.Aliases = hl7.Aliases{"value": "X-2"}
	config, _ := json.Marshal(map[string]*hl7.Modification{"input": input, "output": output})
	if err := os.WriteFile(filepath.Join(dir, "config.json"), config, 0600); err != nil {
		t.Fatal(err)
	}
	str := func(s string) *string { return &s }
	base := hl7.MappingCase{Name: "case", Config: "config.json", InputBlock: "input", OutputBlock: "output", Input: str("A|one"), ExpectedOutput: str("X|one"), ExpectedAliases: map[string]string{"value": "one"}}
	run := func(c hl7.MappingCase) ([]hl7.MappingCaseResult, error) {
		data, _ := json.Marshal(hl7.MappingSuite{Version: 1, Cases: []hl7.MappingCase{c}})
		file := filepath.Join(dir, "suite.json")
		if err := os.WriteFile(file, data, 0600); err != nil {
			t.Fatal(err)
		}
		return hl7.CheckMappings(file)
	}
	got, err := run(base)
	if err != nil || !got[0].Passed {
		t.Fatalf("%v %v", got, err)
	}
	c := base
	c.ExpectedOutput = str("X|two")
	got, err = run(c)
	if err != nil || got[0].Passed || !strings.Contains(got[0].Detail, "byte 2") {
		t.Fatalf("%v %v", got, err)
	}
	c = base
	c.ExpectedAliases = map[string]string{"value": "wrong"}
	got, err = run(c)
	if err != nil || got[0].Passed {
		t.Fatalf("%v %v", got, err)
	}
	c = base
	c.ExpectedOutput = nil
	if _, err := run(c); err == nil {
		t.Fatal("missing expected output accepted")
	}
	c = base
	c.InputFile = "missing.txt"
	if _, err := run(c); err == nil {
		t.Fatal("two input sources accepted")
	}
	c = base
	c.ExpectedOutput = nil
	c.ExpectedAliases = nil
	c.ExpectedError = "undefined_input_tag"
	c.Input = str("UNKNOWN|one")
	got, err = run(c)
	if err != nil || !got[0].Passed {
		t.Fatalf("%v %v", got, err)
	}
	c.ExpectedError = "parse_error"
	got, err = run(c)
	if err != nil || got[0].Passed {
		t.Fatalf("%v %v", got, err)
	}
	c = base
	c.Input = nil
	c.InputFile = "input.txt"
	c.ExpectedOutput = nil
	c.ExpectedOutputFile = "expected.txt"
	_ = os.WriteFile(filepath.Join(dir, "input.txt"), []byte("A|one"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "expected.txt"), []byte("X|one"), 0600)
	got, err = run(c)
	if err != nil || !got[0].Passed {
		t.Fatalf("%v %v", got, err)
	}
	file := filepath.Join(dir, "suite.json")
	for _, data := range []string{`{"version":1,"cases":[],"typo":true}`, `{"version":1,"cases":[]} {}`, `{"version":2,"cases":[]}`} {
		_ = os.WriteFile(file, []byte(data), 0600)
		if _, err := hl7.CheckMappings(file); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
}

func TestExplainAndSafeDiagnostics(t *testing.T) {
	input, output := auditMods("|")
	output.Aliases = hl7.Aliases{"first": "X-2"}
	plan, err := hl7.ExplainMapping(&hl7.ConverterParams{InputModification: input, OutputModification: output}, true)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Mode != "positions" || len(plan.Tags) != 1 || plan.Tags[0].Fields[1].Links[0] != "A-2" {
		t.Fatalf("%+v", plan)
	}
	plan.Aliases["first"] = "changed"
	if output.Aliases["first"] != "X-2" {
		t.Fatal("explain mutated config")
	}
	cause := hl7.NewErrUndefinedInputTag("PATIENT-SECRET", "PATIENT-SECRET")
	diagnostic := hl7.Diagnose(cause)
	data, _ := json.Marshal(diagnostic)
	if diagnostic.Code != "undefined_input_tag" || strings.Contains(string(data), "PATIENT-SECRET") {
		t.Fatal(string(data))
	}
	if hl7.ErrorCode(hl7.NewError(hl7.ErrInvalidConfig, false)) != "invalid_config" {
		t.Fatal("code lost")
	}
	if _, err := hl7.ExplainMapping(nil, false); !errors.Is(err, hl7.ErrInvalidConfig) {
		t.Fatal(err)
	}
}

func ExampleCheckMappings() {
	results, err := hl7.CheckMappings("examples/mappings/suite.json")
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		fmt.Printf("%s: %t\n", result.Name, result.Passed)
	}
	// Output:
	// repeated rows and aliases: true
	// input order: true
	// unknown tag rejected: true
}

func TestExplainDefaultsAndParseOnlyMetadata(t *testing.T) {
	input, output := auditMods("|")
	tag := output.TagsInfo.Tags["X"]
	tag.Tempalate = "X|??constant"
	output.TagsInfo.Tags["X"] = tag
	params := &hl7.ConverterParams{InputModification: input, OutputModification: output}
	plan, err := hl7.ExplainMapping(params, false)
	if err != nil || plan.Tags[0].Fields[1].Default == nil || *plan.Tags[0].Fields[1].Default != "constant" {
		t.Fatalf("%+v %v", plan, err)
	}
	tag.Tempalate = ""
	output.TagsInfo.Tags["X"] = tag
	c := auditConverter(t, input, output)
	parsed, err := c.ParseInput([]byte("A|one"))
	if err != nil || parsed.String() != "A|one" {
		t.Fatalf("%v %v", parsed, err)
	}
	if _, err := c.Convert([]byte("A|one")); !errors.Is(err, hl7.ErrInvalidConfig) {
		t.Fatal(err)
	}
}

func TestFixtureEmptyAliasExpectationRoundTrip(t *testing.T) {
	empty := map[string]string{}
	encoded, err := json.Marshal(hl7.MappingCase{ExpectedAliases: empty})
	if err != nil {
		t.Fatal(err)
	}
	var decoded hl7.MappingCase
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ExpectedAliases == nil {
		t.Fatal("explicit empty alias expectation was lost")
	}
}
