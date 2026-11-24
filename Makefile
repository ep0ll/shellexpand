.PHONY: all test race vet lint fmt fmt-check check pre-commit verify install-hooks clean

all: check

test:
	go test -count=1 ./...

race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run --timeout=5m

fmt:
	gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')
	golangci-lint fmt

fmt-check:
	@test -z "$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'))" || \
		(echo "Go files need gofmt; run 'make fmt'." >&2; exit 1)

check: fmt-check test race vet lint

pre-commit: fmt-check test vet

verify: check

install-hooks:
	./scripts/install-hooks.sh

clean:
	rm -f coverage.out coverage.html
	rm -rf dist/ tmp/
