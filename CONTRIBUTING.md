# Contributing

## Setup

Use the Nix dev shell when possible:

```bash
nix develop
```

This provides Go 1.26, Terraform, `golangci-lint`, Node, `pnpm`, `pnpx`, and Docker Compose. If you do not use Nix, install those tools locally.

## Checks

Run the same checks used by pull request CI:

```bash
make ci
```

Individual targets are available while iterating:

```bash
make fmt
make lint
make test
make build
```

Integration tests require LocalStack:

```bash
make localstack-up
make test-integration
make localstack-down
```

## Changesets

For user-visible changes, add a changeset:

```bash
pnpx changeset
```

Documentation-only and CI-only changes do not need a changeset. The release workflow creates a release PR from merged changesets. Merging that release PR updates `CHANGELOG.md`, creates the version tag, and publishes GitHub release notes.

## Pull Requests

Before opening a pull request, run `make ci`. PR CI runs formatting, vet, linting, unit tests, build checks, and a Changesets status check for code changes. Terraform/LocalStack integration tests run on pushes to `main`, manually, and nightly.
