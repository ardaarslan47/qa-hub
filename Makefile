PLAYWRIGHT_INSTALL_FLAGS ?=

.PHONY: check generate vet test build e2e

check: generate vet test build e2e

generate:
	go tool templ generate

vet: generate
	go vet ./...

test: generate
	go test ./...

build: generate
	go build -o bin/qa-hub ./cmd/qa-hub

e2e: build
	cd e2e && npm ci && npx playwright install $(PLAYWRIGHT_INSTALL_FLAGS) chromium && npm run typecheck && npm test
