# Contributing to easySFTP

Thanks for your interest in contributing! Bug reports, documentation fixes and
features are all welcome. This document tells you everything you need to get a
change merged.

Please note that this project has a [Code of Conduct](CODE_OF_CONDUCT.md);
by participating you agree to abide by it.

## Reporting bugs & requesting features

- Search the [existing issues](https://github.com/eiserv/easySFTP/issues) first.
- **Bugs:** include your workflow step (redact secrets!), the action version,
  the runner OS and the relevant log output. A `dry-run: true` log is often
  enough to reproduce planning problems.
- **Security vulnerabilities:** do **not** open a public issue, see
  [SECURITY.md](SECURITY.md).
- **Features:** describe the use case, not just the solution. Issues labeled
  [`good first issue`](https://github.com/eiserv/easySFTP/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)
  are a great place to start; comment on an issue before starting bigger work
  so nobody duplicates effort.

## Development setup

You need [Go](https://go.dev/dl/) (version from [go.mod](go.mod)), nothing
else. No Docker required for tests.

```console
$ git clone https://github.com/eiserv/easySFTP.git
$ cd easySFTP
$ go test ./...            # unit + end-to-end tests (in-process SFTP server)
$ go vet ./...
$ gofmt -l .               # must print nothing
$ go build ./cmd/easysftp
```

### Repository layout

```
action.yml               the composite action: inputs → EASYSFTP_* env vars
cmd/easysftp/            binary entry point the action runs
cmd/easysftp-bench/      benchmark harness: measures a run and files it under benchmarks/
cmd/linkprobe/           prints the network-path measurement (see internal/linkprobe) as JSON
internal/config/         env + YAML config parsing and validation
internal/uploader/       SFTP connection, planning, strategies, transfers
internal/autotune/       resolves transport settings left at "auto": connections,
                         parallel files, requests in flight
internal/autocache/      carries what one run measured about its server to the next
internal/metrics/        benchmark instrumentation for one run; off unless a
                         metrics file is named
internal/linkprobe/      measures the network path to the SFTP server: round-trip
                         time, a control throughput, server load
internal/gha/            GitHub Actions helpers (outputs, annotations, summary)
internal/actionmeta/     drift-check tests over action.yml and this guide
internal/benchmark/      the measurement harness (8 sub-packages: driver, link,
                         report, runner, scenario, schema, stats, store)
schema/                  JSON Schema for the YAML config file
benchmarks/              stored benchmark results, filed by cmd/easysftp-bench
scripts/                 CI helper scripts the action and its self-test use
                         (prepare-action.sh, test-action.sh, action-lib.sh,
                         mask-credentials.sh)
site/                    the project website (static HTML)
docs/                    user documentation
```

`internal/benchmark/`, `internal/linkprobe`, `internal/metrics`,
`cmd/easysftp-bench`, `cmd/linkprobe` and the stored results in `benchmarks/`
are the benchmark harness. Together they are about half the Go code in this
repository, and none of them is needed to work on the action itself: start in
`cmd/easysftp`, `internal/config` and `internal/uploader`.

### Running the binary locally

The binary is configured entirely through `EASYSFTP_*` environment variables.
See [action.yml](action.yml) for the mapping. Example against a local SFTP
server:

```console
$ EASYSFTP_HOST=localhost EASYSFTP_PORT=2222 \
  EASYSFTP_USERNAME=demo EASYSFTP_PASSWORD=demopass \
  EASYSFTP_ALLOW_ANY_HOST_KEY=true \
  EASYSFTP_SOURCE=./dist EASYSFTP_TARGET=/upload \
  EASYSFTP_DRY_RUN=true \
  go run ./cmd/easysftp
```

`EASYSFTP_ALLOW_ANY_HOST_KEY=true` is only acceptable against a throwaway
local server. Against anything else, pin the host key with
`EASYSFTP_HOST_KEY` or `EASYSFTP_KNOWN_HOSTS`; in v3 an unverified
connection is never the silent default.

Every `EASYSFTP_*` variable this guide names must be a live v3 input: a
removed v2 variable does not misbehave, it fails the run with a migration
error before anything happens (see
[internal/config/config.go](internal/config/config.go), `removedInputs`).
[internal/actionmeta](internal/actionmeta) holds the tests that enforce this
and that load the example above through the real config parser, so neither
can drift back into v2 inputs.

### Tests

- Unit and end-to-end tests run against an **in-process SFTP server**
  ([internal/uploader/testserver_test.go](internal/uploader/testserver_test.go)),
  so `go test ./...` needs no network and no Docker.
- CI additionally runs the whole action against a real OpenSSH server
  ([.github/workflows/ci.yml](.github/workflows/ci.yml)) on Linux and runs the
  unit tests on Windows.
- New behavior needs a test. Bug fixes need a test that fails without the fix.

## Pull requests

1. Fork, create a branch, make your change.
2. Make sure `go test ./...`, `go vet ./...` and `gofmt -l .` are clean.
3. Open a PR against `main`.

### PR title = Conventional Commit (required)

PRs are **squash-merged**, so the PR title becomes the commit message and must
be a [Conventional Commit](https://www.conventionalcommits.org/). CI enforces
this. The prefix decides the release bump
(see [docs/RELEASING.md](docs/RELEASING.md)):

| Prefix | Effect | Example |
|---|---|---|
| `fix:` | patch release | `fix: retry on transient EOF` |
| `feat:` | minor release | `feat: add keepalive support` |
| `feat!:` / `BREAKING CHANGE:` | major release | `feat!: fail without host key` |
| `docs:` `ci:` `chore:` `refactor:` `test:` `build:` `perf:` | no release | `docs: fix typo` |

### Review checklist (what the maintainer looks for)

- Behavior changes are covered by tests and documented (README / `docs/` /
  `action.yml` input descriptions).
- No new dependencies unless truly needed.
- Anything CI pulls from outside the repo is pinned by hash: actions by their
  full commit SHA, container images by `@sha256:` digest (with the readable
  tag kept as a trailing comment). Dependabot does not touch inline
  `docker run` digests, so those are bumped by hand.
- Errors are wrapped with context (`fmt.Errorf("...: %w", err)`), log output
  goes through the `Logger`/`gha` helpers.
- Destructive-path changes (anything that deletes remote files) keep the
  [delete guards](docs/strategies.md#delete-guards) intact.

Releases are fully automated: you never bump versions or edit the changelog
by hand. Merged `feat:`/`fix:` PRs ship with the next release PR merge.
