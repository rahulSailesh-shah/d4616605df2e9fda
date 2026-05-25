.PHONY: agent test tidy

agent:
	go run ./cmd/agent

test:
	go test ./...

tidy:
	go mod tidy
