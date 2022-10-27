test:
	go clean -testcache ./...
	go test -v ./...
	go test -v ./... -tags=integration

demo:
	go run ./demo/main.go

fmt:
	go fmt ./...
