# industrial-cli-kit

Shared, protocol-agnostic foundations for the `*-cli` industrial protocol tools.

It intentionally contains no protocol client, protocol data model, or
protocol-specific configuration schema. See [CONTRACT.md](CONTRACT.md) for the
public command-line contract.

## Packages

- `command`: command registry and global-flag normalisation
- `completion`: Bash and Zsh completion generation from a registry
- `contracttest`: reusable baseline command-contract checks
- `exitcode`: stable process exit statuses
- `output`: snapshot and stream output helpers
- `safety`: dry-run and `--yes` confirmation policy

## Development

```sh
go test ./...
go vet ./...
```
