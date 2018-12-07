GOFILES := $(shell find . -path ./vendor -prune -o -type f -name '*.go' -print)

all: test

test:
	go test ./...

coverage: coverage.txt

coverage.txt: $(GOFILES) Makefile
	go test -coverprofile=coverage.txt -covermode=atomic ./...

show-coverage: coverage.txt
	go tool cover -html=coverage.txt

style: $(GOFILES) Makefile .gometalinter.json
	bin/style

clean:
	go clean
	rm -f coverage.txt

.PHONY: clean test coverage show-coverage
