.PHONY: run build test vet lint check demo

run:
	go run ./cmd/pos

build:
	go build ./...

test:
	go test -count=1 -cover ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

check: build test vet
	@files=$$(gofmt -l cmd internal) && { test -z "$$files" || { echo "$$files"; exit 1; }; }

demo:
	@set -eu; dir=$$(mktemp -d); trap 'rm -rf "$$dir"' EXIT; \
	go build -o "$$dir/pos" ./cmd/pos; \
	(cd "$$dir" && ./pos) < examples/demo.txt; \
	cat "$$dir/totals.csv"
