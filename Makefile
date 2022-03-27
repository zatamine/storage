.PHONY: test lint clean

GO_VERSION	?= 1.18
GO_CMD			:= go$(GO_VERSION)

test: clean
	${GO_CMD} test -v ./... -cover

lint:
	${GO_CMD} vet ./...

clean:
	${GO_CMD} mod tidy
