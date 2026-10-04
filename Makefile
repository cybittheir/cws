run:
	go run ./cmd/server
fmt:
	gofmt -w ./cmd ./internal
test:
	go test ./...
