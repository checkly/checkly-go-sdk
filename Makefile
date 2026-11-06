test: test-unit test-integration

test-unit:
	go test -count=1 -v ./...

test-integration:
	go test -count=1 -v -tags=integration ./...

demo:
	go run ./demo/main.go

fmt:
	go fmt ./...
