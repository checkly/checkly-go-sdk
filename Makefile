test:
	go clean -testcache ./...
	go test -v ./... -tags=integration

demo:
	go run ./demo/main.go
