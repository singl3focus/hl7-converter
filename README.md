<p align="center">
<a href="https://github.com/singl3focus/hl7-converter/actions/workflows/go.yml"><img src="https://github.com/singl3focus/hl7-converter/actions/workflows/go.yml/badge.svg" alt="CI"></a> <img src="https://img.shields.io/badge/made_by-singl3focus-blue" alt="Made by singl3focus"> <img src="https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat" alt="PRs welcome">
</p>

# HL7 Converter

Go toolkit for converting HL7/ASTM-style lab messages by declarative JSON mappings. Suitable for LIS gateways, ETL pipelines, or integration bridges.

**Highlights**
- Declarative mappings in JSON (validated by schema + runtime checks)
- JS post-processing hook on the Result (Otto)
- Examples, tests, and benchmarks included

## What it is
HL7 Converter maps inbound laboratory/clinical messages to outbound formats using JSON-defined modifications. You describe separators, tag ordering, templates, and aliases; the library parses input, applies templates/links, and produces an output message. No custom Go code is required for each mapping—config drives the transformation.

## When to use
- Building LIS connectors or vendor-to-vendor lab bridges (HL7 ORU/OML, ASTM devices, proprietary row-based feeds).
- ETL pipelines where upstream produces row-delimited records and downstream expects HL7-like layout.
- Rapid prototyping of device integrations without modifying business code—swap configs to add a new mapping.
- Post-processing with small JS snippets to adjust payloads dynamically.

**Compatibility**: ASCII messages only. Uses SemVer; v1+ promises backward-compatible API.

## Quick Start
```bash
go get github.com/singl3focus/hl7-converter/v2@latest
```

```go
package main

import (
    "log"

    hl7converter "github.com/singl3focus/hl7-converter/v2"
)

func main() {
    cfgPath := "./path/to/your/config.json"

    params, err := hl7converter.NewConverterParams(cfgPath, "astm_hbl", "mindray_hbl")
    if err != nil {
        log.Fatal(err)
    }

    conv, err := hl7converter.NewConverter(params, hl7converter.WithUsingPositions(), hl7converter.WithUsingAliases())
    if err != nil {
        log.Fatal(err)
    }

    input := []byte("H|\\^&|||sireAstmCom|||||||P|LIS02-A2|20220327\nP|1||||^||||||||||||||||||||||||||||\nO|1|142212||^^^Urina4^screening^|||||||||^||URI^^||||||||||F|||||\nR|1|^^^Urina4^screening^^tempo-analisi-minuti|180|||||F|||||\nR|2|^^^Urina4^screening^^tempo-analisi-minuti|90|||||F|||||\nL|1|N")

    msgType, err := hl7converter.IndetifyMsg(params, input)
    if err != nil {
        log.Fatal(err)
    }

    result, err := conv.Convert(input)
    if err != nil {
        log.Fatal(err)
    }

    _ = msgType // use message type for routing if needed
    log.Print(result.String())
}
```

Create the JSON mapping file in your project and pass its path to `NewConverterParams`. A repository sample lives in `examples/config.json`.

## Stable API (v1)
- Construction: `NewConverterParams`, `NewConverter` (options: `WithUsingPositions`, `WithUsingAliases`).
- Execution: `Converter.Convert`, `Converter.ParseMsg`, `Converter.ParseInput`.
- Types: `Result` (+ `UseScript`, `Aliases`, `FindTag`, `SwapRows/SetRow`), `Row`, `Field`, `Msg`.
- Config: `ReadJSONConfigBlock`, `ValidateJSONConfig`, `Modification.Validate`.
- Utilities: `IndetifyMsg`, constant `CfgSchemaJSON`.
API is considered stable: no breaking changes without a major version bump; thread safety is not guaranteed (see below).

## Config Guide (schema-backed)
- See config.schema.json; a repository sample config lives in `examples/config.json`.
- Required per modification: component_separator, component_array_separator, field_separator, line_separator, tags_info.
- tags_info.positions describes output order; keys are numeric strings, values are tag names present in tags_info.tags.
- Each tag requires linked, fields_number (use -1 to skip length check), template; optional options currently supports autofill.
- Templates support <TAG-INDEX> (float for components, e.g., <O-16.1>), defaults via ??default.

## How it works
1) Load two modifications (input/output) from JSON.
2) Parse input message into tags/fields using input separators and options (e.g., autofill).
3) Convert either by scanning input tags (default) or by output positions (UsingPositions).
4) Fill templates: literals stay, links like `<TAG-3>` pull fields, `<TAG-3.1>` pulls components, `??value` sets defaults.
5) Optionally apply aliases to expose commonly used values.
6) Optionally run JS over the Result to tweak payload.

## Capabilities
- HL7/ASTM-style row parsing with custom separators.
- Positional or input-driven generation of output rows.
- Component addressing with float-style indexes for components.
- Aliases to surface key values for routing or logging.
- JS post-processing (trusted code) on the full Result object.
- JSON Schema validation plus runtime validation for templates and options.

## Limitations
- ASCII-only payloads are expected.
- No built-in sandbox or timeout for JS (run only trusted scripts).
- Converter/Result are not goroutine-safe; use per-request instances or external sync.

## JS scripts
Result.UseScript exposes the Result as global msg inside an Otto VM. Scripts are trusted (no sandbox/timeout), so run only controlled code.

## Thread safety
Converter and Result are not goroutine-safe yet; do not share instances across goroutines without external synchronization.

## Testing and benchmarks
Run `go test ./...`. Benchmarks live in benchmarks/ (see benchmarks_test.go).

## Practical patterns
- Routing by message type: use `IndetifyMsg` after parsing to branch by type.
- Aliases for common fields: call `ApplyAliases` to extract values like patient ID or header keys without re-walking rows.
- Positional outputs: enable `WithUsingPositions` when the output order is fixed and must include repeated tags with known counts.
- Template defaults: prefer explicit defaults via `??value` to keep failures visible when data is missing.
- JS tweaks: small, deterministic scripts (e.g., renaming a tag or swapping fields) instead of complex business logic.

## Benchmarking
Benchmarks in `benchmarks/` cover realistic device-like payloads. To run: `go test -bench=. ./...`. Use them to gauge performance after changing configs or templates.

## License
MIT
