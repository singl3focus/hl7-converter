// Command hl7conv is a small CLI around the hl7-converter library.
//
// Subcommands:
//
//	hl7conv convert         convert a message using the given config
//	hl7conv identify        identify a message type using the given config
//	hl7conv validate-config validate a JSON config and optionally a modification block
//	hl7conv version         print build information
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"

	hl7converter "github.com/singl3focus/hl7-converter/v2"
)

// version is overridden at release time via -ldflags by GoReleaser.
var version = "dev"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		usage(stderr)
		return errors.New("no command specified")
	}

	switch args[0] {
	case "convert":
		return cmdConvert(args[1:], stdin, stdout)
	case "identify":
		return cmdIdentify(args[1:], stdin, stdout)
	case "validate-config":
		return cmdValidateConfig(args[1:], stdout)
	case "version":
		return cmdVersion(stdout)
	case "-h", "--help", "help":
		usage(stdout)
		return nil
	default:
		usage(stderr)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, `hl7conv - convert and inspect row-based lab messages with hl7-converter.

Usage:
  hl7conv <command> [flags]

Commands:
  convert         Convert a message using a config and an input/output block pair.
  identify        Identify a message type by matching parsed tags against configured Types.
  validate-config Validate a JSON config file and optionally a single modification block.
  version         Print version.

Run 'hl7conv <command> -h' for command-specific flags.`)
}

func cmdVersion(w io.Writer) error {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			fmt.Fprintln(w, info.Main.Version)
			return nil
		}
	}
	fmt.Fprintln(w, version)
	return nil
}

func cmdConvert(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := fs.String("config", "", "path to JSON config file (required)")
	in := fs.String("in", "", "input modification block name (required)")
	out := fs.String("out", "", "output modification block name (required)")
	input := fs.String("input", "", "input message file; when empty, read from stdin")
	usePositions := fs.Bool("positions", false, "use position-driven conversion")
	useAliases := fs.Bool("aliases", false, "apply aliases after conversion")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *cfg == "" || *in == "" || *out == "" {
		return errors.New("convert: -config, -in and -out are required")
	}

	params, err := hl7converter.NewConverterParams(*cfg, *in, *out)
	if err != nil {
		return err
	}

	opts := make([]hl7converter.OptionFunc, 0, 2)
	if *usePositions {
		opts = append(opts, hl7converter.WithUsingPositions())
	}
	if *useAliases {
		opts = append(opts, hl7converter.WithUsingAliases())
	}

	conv, err := hl7converter.NewConverter(params, opts...)
	if err != nil {
		return err
	}

	msg, err := readInput(*input, stdin)
	if err != nil {
		return err
	}

	result, err := conv.Convert(msg)
	if err != nil {
		return err
	}

	if _, err := stdout.Write(result.Bytes()); err != nil {
		return err
	}
	if len(result.Bytes()) == 0 || result.Bytes()[len(result.Bytes())-1] != '\n' {
		fmt.Fprintln(stdout)
	}

	return nil
}

func cmdIdentify(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := flag.NewFlagSet("identify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := fs.String("config", "", "path to JSON config file (required)")
	in := fs.String("in", "", "input modification block name (required)")
	out := fs.String("out", "", "output modification block name (required); identify still needs a valid pair")
	input := fs.String("input", "", "input message file; when empty, read from stdin")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *cfg == "" || *in == "" || *out == "" {
		return errors.New("identify: -config, -in and -out are required")
	}

	params, err := hl7converter.NewConverterParams(*cfg, *in, *out)
	if err != nil {
		return err
	}

	msg, err := readInput(*input, stdin)
	if err != nil {
		return err
	}

	msgType, err := hl7converter.IndetifyMsg(params, msg)
	if err != nil {
		return err
	}

	fmt.Fprintln(stdout, msgType)
	return nil
}

func cmdValidateConfig(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("validate-config", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := fs.String("config", "", "path to JSON config file (required)")
	block := fs.String("block", "", "optional modification block name to fully validate")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *cfg == "" {
		return errors.New("validate-config: -config is required")
	}

	ok, err := hl7converter.ValidateJSONConfig(*cfg)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("config is not valid against the embedded JSON schema")
	}

	fmt.Fprintln(stdout, "schema: ok")

	if *block != "" {
		mod, err := hl7converter.ReadJSONConfigBlock(*cfg, *block)
		if err != nil {
			return err
		}
		if err := mod.Validate(); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "block %q: ok\n", *block)
	}

	return nil
}

func readInput(path string, stdin io.Reader) ([]byte, error) {
	if path == "" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}
