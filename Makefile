.PHONY: check generate vet test build e2e run

check: vet test e2e

generate:
	go tool templ generate

vet: generate
	go vet ./...

test: generate
	go test ./...

build: generate
	go build -o bin/qa-hub ./cmd/qa-hub

e2e: e2e/node_modules/.package-lock.json
	cd e2e && npx playwright test

e2e/node_modules/.package-lock.json: e2e/package-lock.json
	cd e2e && npm ci
	cd e2e && npx playwright install chromium

run: build
	./bin/qa-hub
