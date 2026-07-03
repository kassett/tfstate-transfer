# tfstate-transfer

`tfstate-transfer` is a CLI for moving Terraform resources from one state to another. It imports matching resources into a target Terraform directory and removes successfully transferred top-level resources from the source state.

## Installation

Install the latest tagged version with Go:

```bash
go install github.com/kassett/tfstate-transfer@latest
```

Or build from this repository:

```bash
go build ./...
```

## Usage

Transfer resources by passing a source directory, a target directory, and one or more resource names:

```bash
tfstate-transfer \
  --source-dir ./source \
  --target-dir ./target \
  --r module.db
```

If the target resource has a different address, separate the source and target with a colon:

```bash
tfstate-transfer \
  --source-dir ./source \
  --target-dir ./target \
  --r module.db_source:module.db_target
```

Preview Terraform commands without importing or removing state:

```bash
tfstate-transfer \
  --source-dir ./source \
  --target-dir ./target \
  --r module.db \
  --dry-run
```

You can also pass a JSON config file:

```bash
tfstate-transfer --config-file ./transfer.json
```

```json
{
  "sourceDir": "./source",
  "targetDir": "./target",
  "resources": [
    {
      "source": "module.db_source",
      "target": "module.db_target"
    }
  ]
}
```

## Development

The recommended development environment uses Nix:

```bash
nix develop
make ci
```

The dev shell provides Go, Terraform, `golangci-lint`, Node, `pnpm`, `pnpx`, and Docker Compose. Without Nix, install those tools locally and run the same Make targets.

Common commands:

```bash
make fmt
make lint
make test
make build
make ci
```

Integration tests require Docker and LocalStack:

```bash
make localstack-up
make test-integration
make localstack-down
```

`make test` runs fast unit tests only. `make test-integration` runs Terraform/LocalStack tests with the `integration` build tag.

## Releases

This repo uses Changesets for changelog entries, release PRs, tags, and GitHub release notes. Add a changeset for user-visible code changes:

```bash
pnpx changeset
```

CI uses the committed `pnpm-lock.yaml` and runs Changesets with `pnpm`. Merging the generated Changesets release PR updates `CHANGELOG.md`, tags the release, and creates GitHub release notes. Releases do not publish an npm package.

## Example Flow

1. Create the new Terraform environment.
2. Copy the desired resources to the target `.tf` files. Do not apply the target configuration first.
3. Run `tfstate-transfer` with the resources to move.
4. Manually import any resources that Terraform reports as unsupported or ambiguous.

