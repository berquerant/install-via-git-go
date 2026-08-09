# AGENTS.md

## Project Overview

`install-via-git-go` is a CLI tool that installs and manages tools via git repositories.
Based on a YAML configuration file, it combines git operations (clone / checkout / pull) with arbitrary shell scripts (check / setup / install / rollback / skip / uninstall) to automate tool installation.

## Repository Layout

```
.
├── main.go              # Entry point
├── main_test.go         # E2E tests (TestEndToEnd)
├── cmd/                 # cobra subcommand definitions
│   ├── root.go          # Root command, common flags, shared resources
│   ├── run.go           # run subcommand (execute installation)
│   ├── uninstall.go     # uninstall subcommand
│   ├── parse.go         # parse subcommand (validate config file)
│   ├── skeleton.go      # skeleton subcommand (generate config template)
│   └── version.go       # version subcommand
├── config/              # Config file structs and parser
├── strategy/            # Installation strategy determination logic
├── runner/              # Installation execution based on strategy
├── git/                 # git CLI wrapper
├── gitlock/             # Lock file (commit hash) management
├── execx/               # Shell script execution utilities
├── errorx/              # Error wrapping utilities
├── exit/                # Process exit utilities
├── filepathx/           # File path abstraction
├── logx/                # Structured logging (zap-based)
├── inspect/             # Inspection utilities
├── lock/                # Lock state type definitions
├── backup/              # Backup for rollback
├── version/             # Version info (embedded at build time)
└── script/              # Script execution related
```

## Architecture

### Installation Flow (`run` subcommand)

1. Check the lock file and local repository state
2. Determine the **Update Strategy** (`strategy` package)
3. Execute `check` scripts (cancels installation on failure)
4. Execute `setup` scripts
5. Manage the local repository (clone / checkout / pull)
6. If no update is required, execute `skip` and exit
7. Execute `install` scripts

If an error occurs at step 5 or later, the lock file and repository are rolled back and `rollback` scripts are executed.

### Update Strategy (`strategy` package)

The strategy is determined from a combination of the following states:

| Factor | Description |
|--------|-------------|
| Local repository existence | `RepoExistence` |
| Lock file existence | `LockExistence` |
| Repository and lock status | `RepoStatus` |
| CLI options | `--update`, `--retry` |
| Subcommand | `run` / `uninstall` |

Key strategy types (`strategy.Type`):

| Strategy | Description |
|----------|-------------|
| `TinitFromEmpty` | Clone repo and create lock |
| `TinitFromEmptyToLock` | Clone repo and checkout to lock |
| `TinitFromEmptyToLatest` | Clone repo and create latest lock |
| `TcreateLock` | Create lock |
| `TcreateLatestLock` | Pull latest and create lock |
| `TupdateToLock` | Checkout to lock |
| `TupdateToLatestWithLock` | Pull latest and update lock |
| `Tnoop` | No operation required |
| `Tretry` | No repo operation, but continue installation |
| `Tnoupdate` | No repo or lock operation, continue installation |
| `Tremove` | Remove the repository |

## Configuration (`config` package)

The config file (default `ivg.yml`) key fields:

```yaml
uri: https://github.com/some/tool.git   # Required
branch: main                             # Default: main
locald: repo                             # Clone destination dir name (default: repo)
lock: lock                               # Lock file name (default: lock)
shell: [/bin/bash]                       # Shell for script execution
env:                                     # Environment variables
  KEY: value
check: [...]                             # Always runs in workDir
setup: [...]                             # Always runs in workDir
install: [...]                           # Runs in workDir/locald
rollback: [...]                          # Runs in workDir/locald
skip: [...]                              # Runs in workDir/locald
uninstall: [...]                         # Runs on uninstall subcommand
```

Environment variables available to scripts:

| Variable | Description |
|----------|-------------|
| `IVG_URI` | Repository URI |
| `IVG_BRANCH` | Branch name |
| `IVG_LOCALD` | Value of `locald` |
| `IVG_LOCK` | Value of `lock` |
| `IVG_WORKD` | Absolute path of workDir (`install` only) |

## Build and Test

```bash
# All tasks (lint + test + build)
go tool task

# Build only
./.github/build.sh -o dist/install-via-git

# Unit tests
go tool gotestsum --format pkgname --format-icons hivis -- -cover -race ./...

# E2E tests (requires network, takes time)
go test -v -run TestEndToEnd .

# Skip E2E tests with short mode
go test -short ./...

# Lint
go vet ./...
go fix -diff ./...

# Code generation (stringer, dataclass, goconfig)
go generate ./...
```

## Coding Conventions

### Language and Style

- **Go 1.26+**
- All changes must pass `go tool task` (lint + test + build)
- Generated files follow the `*_generated.go` naming convention — do **not** edit manually
- `go generate ./...` produces: `stringer` (String() methods), `dataclass` (data classes), `goconfig` (config)

### Error Handling

- Use `errorx.Errorf(err, "context %s", value)` to attach context to errors
- Use `errors.Join` to combine sentinel errors
- Use `exit.Fail()` for fatal errors that should terminate the process

### Logging

- Use the `logx` package
- Format: `logx.Info("message", logx.S("key", value))`
- Use `logx.Debug`, `logx.Info`, `logx.Error` appropriately
- Enable debug logs with the `--debug` flag

### Testing

- Unit tests live in `*_test.go` files within each package
- E2E tests live in `main_test.go` as `TestEndToEnd` (requires network; skipped with `-short`)
- Test framework: `github.com/stretchr/testify`

### Package Design

- Each package has a single responsibility (git ops, lock management, execution, strategy determination, etc.)
- The `cmd` package only defines commands; business logic is delegated to individual packages
- Use `filepathx.Path` / `filepathx.FilePath` / `filepathx.DirPath` for all file path operations

## Release

Multi-platform binaries are released via GoReleaser:

- Target OS: `linux`, `darwin`, `windows`
- Target architectures: `amd64`, `arm64`, `arm` (Linux only)
- Version info is embedded at build time via `ldflags` (references `VERSION_PACKAGE` env var)

## Key Dependencies

| Library | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/goccy/go-yaml` | YAML parser |
| `github.com/berquerant/execx` | Shell script execution |
| `github.com/go-git/go-git/v5` | git operations (transitive) |
| `golang.org/x/exp` | Experimental Go stdlib extensions |
| `gotest.tools/gotestsum` | Test runner |
| `github.com/go-task/task/v3` | Task runner (Taskfile.yml) |
