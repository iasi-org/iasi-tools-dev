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
- `iasi-dev build [-v] [repository]`
- `iasi-dev publish [-v] [repository]`
- `iasi-dev deploy [-v] [-f|--full] [-m|--message "message"] [repository]`
- `iasi-dev commit [-v] -m "message" [repository]`
- `iasi-dev init [options] [workspace]`
- `iasi-dev docker [start|stop|status]`
- `iasi-dev help [command]`

Command output from underlying tools is written to timestamped files under the
workspace `logs` directory; the console only shows IASI status messages.

With `iasi-dev deploy --full`, repositories are processed one by one. IASI Quarto
repositories delegate their incremental build/publish decision to
`iasi.quarto::deploy()` before that repository is committed and pushed.

The `build` and `publish` targets can be nested paths such as `dir1/dir2`.
When supplied, only `dir2` and the IASI Quarto publications below it are
processed; the operation does not expand to the enclosing Git repository.

Add `bin` to `PATH` to invoke `iasi-dev` from any directory. Scripts under `lib`
are internal implementation details and are not part of the public interface.

Internal code is grouped by responsibility under `lib/commands`, `lib/core`,
and `lib/install`. Docker configuration lives at the repository root under
`docker`.
