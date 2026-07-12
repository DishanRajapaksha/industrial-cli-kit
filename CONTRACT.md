# Industrial CLI contract v1

This contract applies to the next compatible release of each participating
protocol CLI. Existing command names remain supported as documented aliases
until the next major version or for at least twelve months, whichever is later.

## Invocation

```text
<tool> [global flags] <command> [command flags] [arguments]
```

Every tool provides `help`/`--help`/`-h`, `version`/`--version`/`-v`,
`init-config`, `validate-config`, `test-connection`, `status`, and
`completions bash|zsh`.

`status` is mandatory. It performs a protocol/device status query when the
protocol supports one; otherwise it is a documented alias for
`test-connection`.

## Global flags

All tools support these global flags before or after a command:

```text
--config PATH --profile NAME --timeout DURATION --format FORMAT --verbose --debug
```

Authentication, TLS, serial, routing, and tracing flags remain
protocol-specific. Where equivalent, use `--endpoint`, `--username`,
`--password`, `--ca-cert`, `--client-cert`, `--client-key`, and
`--insecure-skip-verify`.

## Data operations

`read`, `write`, and `watch` are the canonical data verbs. Multiple reads use
repeated `--item ITEM` and optionally `--items-file PATH`. Writes use
`--item ITEM --type TYPE --value VALUE`.

Mutating operations are dry runs by default. `--yes` permits transmission and
`--dry-run` makes non-transmission explicit; the two flags cannot be combined.

## Output

Snapshot commands accept `table`, `text`, `json`, and `csv`. Streaming commands
accept `text`, `jsonl`, and `csv`. Results are written to stdout; diagnostics
are written to stderr.

New JSON APIs use this envelope, with an RFC 3339 timestamp and a
protocol-specific `result` payload:

```json
{
  "protocol": "modbus",
  "timestamp": "2026-07-12T00:00:00Z",
  "result": {}
}
```

JSONL writes one such result per line. Existing JSON output remains supported
as a compatibility mode through the alias window.

## Exit statuses

| Code | Meaning |
| ---: | --- |
| 0 | success |
| 1 | general/internal failure |
| 2 | usage or configuration failure |
| 3 | connection/transport failure |
| 4 | protocol request failure |
| 5 | authentication/security failure |
| 6 | requested resource missing |
| 7 | rejected operation or missing `--yes` |
| 8 | timeout |
| 9 | output failure |
