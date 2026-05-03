package main

import (
	"bytes"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestCmdValidateConfigOK(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(root, "examples", "config.json")

	var stdout bytes.Buffer
	err := run([]string{"validate-config", "-config", cfg, "-block", "astm_hbl"}, nil, &stdout, &stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, stdout.String())
	}

	if !strings.Contains(stdout.String(), "schema: ok") {
		t.Fatalf("expected schema ok in output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), `block "astm_hbl": ok`) {
		t.Fatalf("expected block ok in output, got %q", stdout.String())
	}
}

func TestCmdConvertFromStdin(t *testing.T) {
	root := repoRoot(t)
	cfg := filepath.Join(root, "examples", "config.json")

	input := []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\n" +
		"P|1||||^||||||||||||||||||||||||||||\n" +
		"O|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\n" +
		"R|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\n" +
		"R|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||")

	var stdout bytes.Buffer
	err := run(
		[]string{"convert", "-config", cfg, "-in", "astm_hbl", "-out", "mindray_hbl"},
		bytes.NewReader(input),
		&stdout,
		&stdout,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, stdout.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "MSH|") || !strings.Contains(out, "OBX|") {
		t.Fatalf("converted output missing expected segments: %q", out)
	}
}

func TestCmdRejectsUnknownCommand(t *testing.T) {
	var stderr bytes.Buffer
	err := run([]string{"nope"}, nil, &stderr, &stderr)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestCmdRequiresFlags(t *testing.T) {
	var stderr bytes.Buffer
	err := run([]string{"convert"}, nil, &stderr, &stderr)
	if err == nil {
		t.Fatal("expected error when required flags are missing")
	}
}
