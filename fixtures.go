package hl7converter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// MappingSuite holds cases with paths relative to the suite file.
type MappingSuite struct {
	Version int           `json:"version"`
	Cases   []MappingCase `json:"cases"`
}

// MappingCase compares exact bytes and optionally aliases or a stable error code.
type MappingCase struct {
	Name               string            `json:"name"`
	Config             string            `json:"config"`
	InputBlock         string            `json:"in"`
	OutputBlock        string            `json:"out"`
	Positions          bool              `json:"positions,omitempty"`
	Input              *string           `json:"input,omitempty"`
	InputFile          string            `json:"input_file,omitempty"`
	ExpectedOutput     *string           `json:"expected_output,omitempty"`
	ExpectedOutputFile string            `json:"expected_output_file,omitempty"`
	ExpectedAliases    map[string]string `json:"expected_aliases"`
	ExpectedError      string            `json:"expected_error,omitempty"`
}

// MappingCaseResult reports a mismatch location without embedding message data.
type MappingCaseResult struct {
	Name      string `json:"name"`
	Passed    bool   `json:"passed"`
	Detail    string `json:"detail,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
}

// CheckMappings runs a version-1 fixture suite. A false Passed value is a case failure.
func CheckMappings(filename string) ([]MappingCaseResult, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var suite MappingSuite
	decoder := json.NewDecoder(f)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&suite); err != nil {
		return nil, fmt.Errorf("fixture: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("fixture must contain exactly one JSON value")
	}
	if suite.Version != 1 || len(suite.Cases) == 0 {
		return nil, fmt.Errorf("fixture requires version 1 and non-empty cases")
	}
	seen := make(map[string]bool)
	for _, c := range suite.Cases {
		if c.Name == "" || seen[c.Name] {
			return nil, fmt.Errorf("fixture case names must be non-empty and unique")
		}
		seen[c.Name] = true
		if c.Config == "" || c.InputBlock == "" || c.OutputBlock == "" {
			return nil, fmt.Errorf("case %q requires config, in and out", c.Name)
		}
		if (c.Input == nil) == (c.InputFile == "") {
			return nil, fmt.Errorf("case %q requires exactly one of input/input_file", c.Name)
		}
		if c.ExpectedError != "" {
			if !knownExpectedError(c.ExpectedError) {
				return nil, fmt.Errorf("case %q: unknown expected_error code %q", c.Name, c.ExpectedError)
			}

			if c.ExpectedOutput != nil || c.ExpectedOutputFile != "" || c.ExpectedAliases != nil {
				return nil, fmt.Errorf("case %q cannot combine expected_error and output/aliases", c.Name)
			}
		} else if (c.ExpectedOutput == nil) == (c.ExpectedOutputFile == "") {
			return nil, fmt.Errorf("case %q requires exactly one expected output", c.Name)
		}
	}
	base := filepath.Dir(filename)
	results := make([]MappingCaseResult, 0, len(suite.Cases))
	for _, c := range suite.Cases {
		result, err := checkMappingCase(base, c)
		if err != nil {
			return nil, fmt.Errorf("case %q: %w", c.Name, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func fixturePath(base, value string) string {
	if filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(base, value)
}

func fixtureBytes(base string, inline *string, filename string) ([]byte, error) {
	if inline != nil {
		return []byte(*inline), nil
	}
	return os.ReadFile(fixturePath(base, filename))
}

func checkMappingCase(base string, c MappingCase) (MappingCaseResult, error) {
	result := MappingCaseResult{Name: c.Name}
	input, err := fixtureBytes(base, c.Input, c.InputFile)
	if err != nil {
		return result, err
	}
	params, err := NewConverterParams(fixturePath(base, c.Config), c.InputBlock, c.OutputBlock)
	var converted *Result
	if err == nil {
		var options []OptionFunc
		if c.Positions {
			options = append(options, WithUsingPositions())
		}
		if c.ExpectedAliases != nil {
			options = append(options, WithUsingAliases())
		}
		var converter *Converter
		converter, err = NewConverter(params, options...)
		if err == nil {
			converted, err = converter.Convert(input)
		}
	}
	result.ErrorCode = ErrorCode(err)
	if c.ExpectedError != "" {
		result.Passed = err != nil && result.ErrorCode == c.ExpectedError
		if !result.Passed {
			result.Detail = fmt.Sprintf("expected error %q, got %q", c.ExpectedError, result.ErrorCode)
		}
		return result, nil
	}
	if err != nil {
		result.Detail = "conversion failed: " + result.ErrorCode
		return result, nil
	}
	expected, err := fixtureBytes(base, c.ExpectedOutput, c.ExpectedOutputFile)
	if err != nil {
		return result, err
	}
	actual := converted.Bytes()
	if string(actual) != string(expected) {
		at := 0
		for at < len(actual) && at < len(expected) && actual[at] == expected[at] {
			at++
		}
		result.Detail = fmt.Sprintf("output differs at byte %d (expected %d bytes, got %d)", at, len(expected), len(actual))
		return result, nil
	}
	if c.ExpectedAliases != nil {
		aliases, _ := converted.Aliases()
		if !reflect.DeepEqual(map[string]string(aliases), c.ExpectedAliases) {
			result.Detail = "aliases differ (expected keys: " + strings.Join(sortedKeys(c.ExpectedAliases), ", ") + ")"
			return result, nil
		}
	}
	result.Passed = true
	return result, nil
}

func knownExpectedError(code string) bool {
	switch code {
	case "invalid_config", "invalid_link", "parse_error", "undefined_input_tag", "output_tag_not_found", "component_error", "row_index_error", "alias_error", "field_count_error":
		return true
	default:
		return false
	}
}
