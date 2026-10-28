// Command hl7conv converts messages and checks declarative mappings.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"sort"
	"strings"

	hl7 "github.com/singl3focus/hl7-converter/v2"
)

var version = "dev"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if jsonRequested(os.Args[1:]) {
			_ = json.NewEncoder(os.Stderr).Encode(hl7.Diagnose(err))
		} else {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
		os.Exit(1)
	}
}

func jsonRequested(args []string) bool {
	enabled := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		switch name {
		case "config", "in", "out", "input", "fixtures", "block":
			if !hasValue {
				i++
			}
		case "json":
			enabled = !hasValue || value == "true"
		}
	}
	return enabled
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		if err := usage(stderr); err != nil {
			return err
		}
		return errors.New("no command specified")
	}
	switch args[0] {
	case "convert":
		return cmdConvert(args[1:], stdin, stdout)
	case "identify":
		return cmdIdentify(args[1:], stdin, stdout)
	case "validate-config":
		return cmdValidateConfig(args[1:], stdout)
	case "check":
		return cmdCheck(args[1:], stdout)
	case "explain":
		return cmdExplain(args[1:], stdout)
	case "version":
		fs := flag.NewFlagSet("version", flag.ContinueOnError)
		if done, err := parseFlags(fs, args[1:], stdout); done || err != nil {
			return err
		}
		if jsonRequested(args[1:]) {
			var value bytes.Buffer
			if err := cmdVersion(&value); err != nil {
				return err
			}
			return json.NewEncoder(stdout).Encode(map[string]string{"version": strings.TrimSpace(value.String())})
		}
		return cmdVersion(stdout)
	case "-h", "--help", "help":
		return usage(stdout)
	default:
		return fmt.Errorf("unknown command %q; run hl7conv -h", args[0])
	}
}

func usage(w io.Writer) error {
	return writeText(w, `hl7conv - convert and verify row-based lab message mappings.

Usage:
  hl7conv <command> [flags]

Commands:
  convert          Convert a message using a config and block pair.
  identify         Identify a message type.
  validate-config  Validate all blocks, one block, or a block pair.
  check            Check exact output and aliases against mapping fixtures.
  explain          Describe output field dependencies without a message.
  version          Print version.

Run 'hl7conv <command> -h' for flags.
Use -json on commands for structured diagnostics on stderr.
`)
}

func cmdVersion(w io.Writer) error {
	value := version
	if value == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			value = info.Main.Version
		}
	}
	return writeText(w, value+"\n")
}

func parseFlags(fs *flag.FlagSet, args []string, stdout io.Writer) (bool, error) {
	var help bytes.Buffer
	fs.SetOutput(&help)
	fs.Bool("json", false, "write structured diagnostics; JSON results for identify/validate/check/explain")
	err := fs.Parse(args)
	if errors.Is(err, flag.ErrHelp) {
		return true, writeBytes(stdout, help.Bytes())
	}
	if err != nil {
		return false, err
	}
	if fs.NArg() != 0 {
		return false, fmt.Errorf("%s: unexpected positional arguments", fs.Name())
	}
	return false, nil
}

func cmdConvert(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	cfg := fs.String("config", "", "path to JSON config (required)")
	in := fs.String("in", "", "input modification (required)")
	out := fs.String("out", "", "output modification (required)")
	input := fs.String("input", "", "message file; default stdin")
	positions := fs.Bool("positions", false, "generate output in configured positions")
	aliases := fs.Bool("aliases", false, "resolve configured output aliases")
	raw := fs.Bool("raw", false, "write exact result bytes without an added LF")
	if done, err := parseFlags(fs, args, stdout); done || err != nil {
		return err
	}
	if *cfg == "" || *in == "" || *out == "" {
		return errors.New("convert: -config, -in and -out are required")
	}
	params, err := hl7.NewConverterParams(*cfg, *in, *out)
	if err != nil {
		return err
	}
	var options []hl7.OptionFunc
	if *positions {
		options = append(options, hl7.WithUsingPositions())
	}
	if *aliases {
		options = append(options, hl7.WithUsingAliases())
	}
	converter, err := hl7.NewConverter(params, options...)
	if err != nil {
		return err
	}
	data, err := readInput(*input, stdin)
	if err != nil {
		return err
	}
	result, err := converter.Convert(data)
	if err != nil {
		return err
	}
	output := result.Bytes()
	if err := writeBytes(stdout, output); err != nil {
		return err
	}
	if !*raw && (len(output) == 0 || output[len(output)-1] != '\n') {
		return writeText(stdout, "\n")
	}
	return nil
}

func cmdIdentify(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("identify", flag.ContinueOnError)
	cfg := fs.String("config", "", "path to JSON config (required)")
	in := fs.String("in", "", "input modification (required)")
	out := fs.String("out", "", "output modification for pair validation (required)")
	input := fs.String("input", "", "message file; default stdin")
	if done, err := parseFlags(fs, args, stdout); done || err != nil {
		return err
	}
	if *cfg == "" || *in == "" || *out == "" {
		return errors.New("identify: -config, -in and -out are required")
	}
	params, err := hl7.NewConverterParams(*cfg, *in, *out)
	if err != nil {
		return err
	}
	data, err := readInput(*input, stdin)
	if err != nil {
		return err
	}
	name, err := hl7.IdentifyMsg(params, data)
	if err != nil {
		return err
	}
	if jsonRequested(args) {
		return json.NewEncoder(stdout).Encode(map[string]string{"type": name})
	}
	return writeText(stdout, name+"\n")
}

func cmdValidateConfig(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("validate-config", flag.ContinueOnError)
	cfg := fs.String("config", "", "path to JSON config (required)")
	block := fs.String("block", "", "validate only one modification")
	in := fs.String("in", "", "validate a conversion pair with -out")
	out := fs.String("out", "", "validate a conversion pair with -in")
	if done, err := parseFlags(fs, args, stdout); done || err != nil {
		return err
	}
	if *cfg == "" {
		return errors.New("validate-config: -config is required")
	}
	if (*in == "") != (*out == "") || (*block != "" && *in != "") {
		return errors.New("use -block or both -in and -out")
	}
	names := []string{}
	pair := ""
	if *in != "" {
		if _, err := hl7.NewConverterParams(*cfg, *in, *out); err != nil {
			return err
		}
		pair = *in + " -> " + *out
	} else if *block != "" {
		if _, err := hl7.ReadJSONConfigBlock(*cfg, *block); err != nil {
			return err
		}
		names = append(names, *block)
	} else {
		blocks, err := hl7.ReadJSONConfig(*cfg)
		if err != nil {
			return err
		}
		for name := range blocks {
			names = append(names, name)
		}
		sort.Strings(names)
	}
	if jsonRequested(args) {
		return json.NewEncoder(stdout).Encode(map[string]any{"schema": "ok", "blocks": names, "pair": pair})
	}
	var text strings.Builder
	text.WriteString("schema: ok\n")
	for _, name := range names {
		fmt.Fprintf(&text, "block %q: ok\n", name)
	}
	if pair != "" {
		fmt.Fprintf(&text, "pair %s: ok\n", pair)
	}
	return writeText(stdout, text.String())
}

func cmdCheck(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fixture := fs.String("fixtures", "", "version-1 mapping suite (required)")
	if done, err := parseFlags(fs, args, stdout); done || err != nil {
		return err
	}
	if *fixture == "" {
		return errors.New("check: -fixtures is required")
	}
	results, err := hl7.CheckMappings(*fixture)
	if err != nil {
		return err
	}
	failed := 0
	for _, result := range results {
		if !result.Passed {
			failed++
		}
	}
	if jsonRequested(args) {
		if err := json.NewEncoder(stdout).Encode(results); err != nil {
			return err
		}
	} else {
		var text strings.Builder
		for _, result := range results {
			status := "PASS"
			if !result.Passed {
				status = "FAIL"
			}
			fmt.Fprintf(&text, "%s %s", status, result.Name)
			if result.Detail != "" {
				fmt.Fprintf(&text, ": %s", result.Detail)
			}
			text.WriteByte('\n')
		}
		if err := writeText(stdout, text.String()); err != nil {
			return err
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d mapping case(s) failed", failed)
	}
	return nil
}

func cmdExplain(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	cfg := fs.String("config", "", "path to JSON config (required)")
	in := fs.String("in", "", "input modification (required)")
	out := fs.String("out", "", "output modification (required)")
	positions := fs.Bool("positions", false, "describe configured position order")
	if done, err := parseFlags(fs, args, stdout); done || err != nil {
		return err
	}
	if *cfg == "" || *in == "" || *out == "" {
		return errors.New("explain: -config, -in and -out are required")
	}
	params, err := hl7.NewConverterParams(*cfg, *in, *out)
	if err != nil {
		return err
	}
	plan, err := hl7.ExplainMapping(params, *positions)
	if err != nil {
		return err
	}
	if jsonRequested(args) {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(plan)
	}
	var text strings.Builder
	fmt.Fprintf(&text, "mode: %s\n", plan.Mode)
	for _, tag := range plan.Tags {
		fmt.Fprintf(&text, "%s linked=%s\n", tag.Tag, tag.Linked)
		for _, field := range tag.Fields {
			fmt.Fprintf(&text, "  %s-%d <- %q\n", tag.Tag, field.Position, field.Expression)
		}
	}
	return writeText(stdout, text.String())
}

func writeBytes(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err == nil && n != len(data) {
		return io.ErrShortWrite
	}
	return err
}

func writeText(w io.Writer, text string) error { return writeBytes(w, []byte(text)) }

func readInput(filename string, stdin io.Reader) ([]byte, error) {
	if filename == "" {
		if stdin == nil {
			return nil, errors.New("stdin is unavailable")
		}
		return io.ReadAll(stdin)
	}
	return os.ReadFile(filename)
}
