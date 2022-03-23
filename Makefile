.PHONY: test lint clean


test:
	go test -v ./... -cover

lint:
	go vet ./...

clean:
	go mod tidy
