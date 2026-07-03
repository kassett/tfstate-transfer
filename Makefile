.PHONY: fmt fmt-check lint test test-integration ci build changeset version localstack-up localstack-down

GO_PACKAGES := ./...
GO_FILES := $$(find . -name '*.go' -not -path './.git/*' -not -path './.cache/*' -not -path './.go/*' -not -path './node_modules/*')

fmt:
	gofmt -w $(GO_FILES)

fmt-check:
	test -z "$$(gofmt -l $(GO_FILES))"

lint:
	go vet $(GO_PACKAGES)
	golangci-lint run

test:
	go test $(GO_PACKAGES) -cover

test-integration:
	go test -tags=integration ./tests -v

build:
	go build ./...

ci: fmt-check lint test build

changeset:
	pnpx changeset

version:
	pnpm changeset version

localstack-up:
	docker compose up -d localstack

localstack-down:
	docker compose down --remove-orphans
