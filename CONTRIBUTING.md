# Contributing to hl7-converter

Thanks for considering a contribution. This project is a small, focused mapping engine, so contributions are most useful when they keep the public API narrow and the conversion semantics deterministic.

## Ground Rules

- The library targets row-based lab messages (HL7, ASTM and similar). Features that pull the project toward a full integration platform (queues, persistence, orchestration) are out of scope.
- Public API stability matters. Prefer additive changes; coordinate breaking changes for the next major.
- Behaviour changes need test coverage. Bug fixes must come with a regression test.

## Development Setup

Minimum compatibility is Go 1.22.3. Use a supported Go release for development and release binaries; CI also checks 1.26.x. The go.mod minimum does not choose the release compiler.

```bash
go mod download
go test ./...
go test -race ./...
go vet ./...
go run ./cmd/hl7conv check -fixtures ./examples/mappings/suite.json
```

Optional benchmark run:

```bash
go test -bench=. ./...
```

## Repository Layout

- `*.go` in the repository root form the public package.
- `examples/` holds the sample config and a runnable usage example.
- `cmd/hl7conv/` is the CLI built on top of the library.
- `benchmarks/` keeps benchmark text reports.
- `internal_helpers.go` holds package-internal helpers and the option registry.

## Coding Guidelines

- Keep changes minimal and focused. Avoid drive-by refactors.
- Do not add new top-level dependencies without prior discussion.
- Errors must wrap a sentinel via the `Error` type so that `errors.Is` keeps working downstream.
- Public symbols need short doc comments.
- Avoid global mutable state; conversion state must remain local to a single `Convert` call.

## Tests

- New behaviour requires unit tests under the existing `*_test.go` files.
- Concurrency-sensitive code should keep `t.Parallel()` and stay race-clean (`go test -race ./...`).
- For new tag options, add coverage in both validation and conversion tests.

## Pull Requests

1. Fork and create a topic branch from `main`.
2. Run the full test and race suites locally.
3. Keep the PR description focused on the user-visible change and migration impact, if any.
4. Reference any relevant issue.

## Releases

Releases are produced manually:

1. Confirm `main` is green.
2. Tag the commit with a `v*` SemVer tag (for example `v2.6.0`).
3. Push the tag. The `release` workflow runs GoReleaser and publishes the GitHub Release.

There is no automated tag bumping. Versioning stays explicit.

## Reporting Issues

Use the issue templates in `.github/ISSUE_TEMPLATE/`:

- bug report
- feature request
- usage question

Include a minimal config, synthetic input and expected output (or a mapping fixture), your current workaround and version/toolchain. Do not publish real patient data.
