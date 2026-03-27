deps:
	go mod download
	go mod verify

test:
	go test -v -race ./...

lint:
	golangci-lint run
