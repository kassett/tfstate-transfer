.PHONY: fmt fmt-check lint test test-integration ci build changeset version localstack-up localstack-down

GO_PACKAGES := ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

fmt-check:
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*'))"

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

