.PHONY: test fmt install

test:
	go test ./...

fmt:
	@command -v goimports >/dev/null 2>&1 || go install golang.org/x/tools/cmd/goimports@latest
	goimports -w -local github.com/everscribe/cli .

install:
	go install ./cmd/es
