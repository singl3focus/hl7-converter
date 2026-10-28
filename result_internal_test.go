package hl7converter

import (
	"errors"
	"testing"
)

func TestResultVMLifecycle(t *testing.T) {
	result := NewResult("\r", []*Row{NewRowWithFieldsValues("|", "^", "~", "X", "one")})
	if result.otto != nil {
		t.Fatal("unused Result allocated a JavaScript VM")
	}
	if err := result.UseScript(nil, ""); !errors.Is(err, ErrScriptNilKey) {
		t.Fatal(err)
	}
	if result.otto != nil {
		t.Fatal("invalid opt-in allocated a VM")
	}
	if err := result.UseScript(KeyScript, "var runs = 1; msg.Rows[0].Fields[1].ChangeValue('two');"); err != nil {
		t.Fatal(err)
	}
	vm := result.otto
	if err := result.UseScript(KeyScript, "runs++; msg.Rows[0].Fields[1].ChangeValue(String(runs));"); err != nil {
		t.Fatal(err)
	}
	if result.otto != vm || result.String() != "X|2" {
		t.Fatal("script state was not reused")
	}
	if err := result.UseScript(KeyScript, "function {"); !errors.Is(err, ErrScriptCompile) {
		t.Fatal(err)
	}
	if err := result.UseScript(KeyScript, "throw new Error('stop');"); !errors.Is(err, ErrScriptRun) {
		t.Fatal(err)
	}
	if err := result.UseScript(KeyScript, "msg.Rows[0].Fields[1].ChangeValue('recovered');"); err != nil {
		t.Fatal(err)
	}
	if result.String() != "X|recovered" {
		t.Fatal(result.String())
	}
}
