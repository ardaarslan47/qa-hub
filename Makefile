.PHONY: check generate vet test build e2e
.NOTPARALLEL:

check: generate vet test build e2e

generate:
	go tool templ generate

vet:
	go vet ./...

test:
	go test ./...

build:
	go build -o bin/qahub ./cmd/qahub

e2e:
	cd e2e && npx playwright test
