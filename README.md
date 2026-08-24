# iasi-tools-dev

Automation scripts and operational utilities for the IASI ecosystem.

## Command-line interface

The development interface is the `iasi-dev` command:

```text
iasi-dev <command> [options]
```

Available commands:

- `iasi-dev clone [options] [workspace]`
- `iasi-dev pull [options] [repository]`
- `iasi-dev sync path [path...]`
- `iasi-dev build [--force] [-v|-s] [repository]`
- `iasi-dev publish [--force] [-v|-s] [repository]`
- `iasi-dev deploy [--force] [-v|-s] [-f|--full] [-m|--message "message"] [repository]`
- `iasi-dev commit [--force] [-v|-s] -m "message" [repository]`
- `iasi-dev init [options] [workspace]`
- `iasi-dev docker [start|stop|status]`
- `iasi-dev help [command]`

Command output from underlying tools is written to timestamped files under the
workspace `logs` directory; the console only shows IASI status messages.

Projects managed by `iasi.quarto` are identified by the presence of `_iasi.yml`.
The `build` and `publish` commands operate only on those projects. `commit` and
`deploy` keep processing all projects recognized by the configured version control
system (VCS). Git is the default and currently implemented VCS.

With `iasi-dev deploy --full`, repositories are processed one by one. Repositories
containing an `_iasi.yml` first delegate their incremental build/publish decision
to `iasi.quarto::deploy()`; repositories without `_iasi.yml` keep the normal VCS
commit/publish flow. `iasi-dev deploy --force` keeps the same project traversal but
forces complete build and publish passes for every IASI Quarto project, ignoring
freshness decisions. If `--full` and `--force` are supplied together, `--force`
takes precedence. In short: `force > full > normal`.

Use `-v` to include detailed operational and success messages. Use `-s` for
silent mode; when both are present, `-s` takes precedence.

Without an explicit repository, `deploy` and `commit` operate on the current
directory. From the `iasi-org` workspace, all direct child projects recognized by the
configured VCS are processed, including `iasi-tools-dev`.

The `build` and `publish` targets can be nested paths such as `dir1/dir2`.
When supplied, only `dir2` and the IASI projects below it are processed; the
operation does not expand to the enclosing Git repository.

Add `bin` to `PATH` to invoke `iasi-dev` from any directory. Scripts under `lib`
are internal implementation details and are not part of the public interface.

Internal code is grouped by responsibility under `lib/commands`, `lib/core`,
and `lib/install`. Docker configuration lives at the repository root under
`docker`.


### Version control systems

`iasi-dev` treats version control as an adapter. Git is the default and currently
implemented VCS:

```bash
IASI_VCS=git
```

The adapter owns the project marker and the version-control operations used by
`iasi-dev`: staging changes, detecting changes, committing, publishing, cloning,
and synchronizing with a remote. The Git adapter lives in `lib/vcs/git.sh`.
Additional VCS implementations can be added as separate adapters without changing
the command orchestration.

Git uses `.git` as its project marker. Each future VCS adapter defines its own
marker and commands. Project discovery only checks the marker supplied by the
active VCS adapter; it does not execute the VCS command.


### Forzar operaciones

`--force` es una opción explícita. En `build`, `publish` y `deploy` fuerza la ejecución completa sin usar decisiones de frescura/estado. En `commit` no se traduce en `push --force` ni crea commits vacíos.
