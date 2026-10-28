package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandHelpAndOutputErrors(t *testing.T) {
	for _, command := range []string{"convert", "identify", "validate-config", "check", "explain", "version"} {
		for _, help := range []string{"-h", "--help"} {
			var output bytes.Buffer
			if err := run([]string{command, help}, nil, &output, io.Discard); err != nil {
				t.Fatalf("%s: %v", command, err)
			}
			if !strings.Contains(output.String(), "Usage") {
				t.Fatal(output.String())
			}
			if err := run([]string{command, help}, nil, failedWriter{}, io.Discard); err == nil {
				t.Fatal("help writer failure lost")
			}
		}
	}
	if err := run([]string{"version"}, nil, failedWriter{}, io.Discard); err == nil {
		t.Fatal("version write failure lost")
	}
	if err := run([]string{"convert", "-unknown"}, nil, io.Discard, io.Discard); err == nil {
		t.Fatal("unknown flag accepted")
	}
	if err := run([]string{"convert", "extra"}, nil, io.Discard, io.Discard); err == nil {
		t.Fatal("extra arg accepted")
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

type secondWriteFails struct{ calls int }

func (w *secondWriteFails) Write(p []byte) (int, error) {
	w.calls++
	if w.calls == 2 {
		return 0, io.ErrClosedPipe
	}
	return len(p), nil
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return 0, nil }

func TestConvertRawAndWrites(t *testing.T) {
	cfg := filepath.Join(repoRoot(t), "examples", "mappings", "config.json")
	args := []string{"convert", "-config", cfg, "-in", "input", "-out", "output"}
	for _, raw := range []bool{false, true} {
		command := append([]string(nil), args...)
		if raw {
			command = append(command, "-raw")
		}
		var output bytes.Buffer
		if err := run(command, strings.NewReader("A|one^two\nA|three^four"), &output, io.Discard); err != nil {
			t.Fatal(err)
		}
		want := "X|one|two\rX|three|four"
		if !raw {
			want += "\n"
		}
		if output.String() != want {
			t.Fatalf("got %q want %q", output.String(), want)
		}
	}
	if err := run(args, strings.NewReader("A|one^two"), &secondWriteFails{}, io.Discard); !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
	if err := run(args, strings.NewReader("A|one^two"), shortWriter{}, io.Discard); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
}

func TestNewCommands(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(root, "examples", "mappings", "config.json")
	commands := [][]string{
		{"validate-config", "-config", cfg},
		{"validate-config", "-config", cfg, "-in", "input", "-out", "output"},
		{"identify", "-config", cfg, "-in", "input", "-out", "output", "-json"},
		{"explain", "-config", cfg, "-in", "input", "-out", "output", "-positions", "-json"},
		{"check", "-fixtures", filepath.Join(root, "examples", "mappings", "suite.json"), "-json"},
	}
	for _, command := range commands {
		var output bytes.Buffer
		if err := run(command, strings.NewReader("A|one^two"), &output, io.Discard); err != nil {
			t.Fatalf("%v: %v", command, err)
		}
		if output.Len() == 0 {
			t.Fatal("empty output")
		}
	}
	if err := run([]string{"validate-config", "-config", cfg, "-in", "input"}, nil, io.Discard, io.Discard); err == nil {
		t.Fatal("partial pair accepted")
	}
	if err := run([]string{"validate-config", "-config", cfg}, nil, failedWriter{}, io.Discard); err == nil {
		t.Fatal("validation write failure lost")
	}
}

func TestValidationSemanticsAndFixtureFailure(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.json")
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "examples", "mappings", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	broken := strings.Replace(string(raw), `"component_separator": "^"`, `"component_separator": "|"`, 1)
	if broken == string(raw) {
		t.Fatal("fixture unchanged")
	}
	if err := os.WriteFile(cfg, []byte(broken), 0600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"validate-config", "-config", cfg}, nil, io.Discard, io.Discard); err == nil {
		t.Fatal("semantic config conflict accepted")
	}
	suite := filepath.Join(dir, "suite.json")
	data := `{"version":1,"cases":[{"name":"wrong","config":"CONFIG","in":"input","out":"output","input":"A|one^two","expected_output":"wrong"}]}`
	// Marshal a path instead of relying on platform-specific slash escaping.
	quoted, _ := json.Marshal(filepath.Join(repoRoot(t), "examples", "mappings", "config.json"))
	data = strings.Replace(data, `"CONFIG"`, string(quoted), 1)
	if err := os.WriteFile(suite, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run([]string{"check", "-fixtures", suite}, nil, &output, io.Discard); err == nil {
		t.Fatal("mismatch returned success")
	}
	if !strings.Contains(output.String(), "FAIL wrong") {
		t.Fatal(output.String())
	}
}

func TestJSONFlagParsing(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{[]string{"convert", "-json", "-json=false"}, false},
		{[]string{"convert", "-json=false", "-json=true"}, true},
		{[]string{"convert", "-input", "-json"}, false},
		{[]string{"convert", "--", "-json"}, false},
		{[]string{"check", "--json"}, true},
	}
	for _, c := range cases {
		if got := jsonRequested(c.args); got != c.want {
			t.Fatalf("%v: %v", c.args, got)
		}
	}
}
