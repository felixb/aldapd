.PHONY: all build clean test

.get-deps: *.go
	go get -t -d -v ./...
	touch .get-deps

all: test build

build: aldapd

clean:
	rm -f .get-deps
	rm -f *_amd64 *_darwin *.exe

test: .get-deps *.go
	go test -v -cover ./...

aldapd: .get-deps *.go
	go build -o $@ *.go

fmt: *.go
	go fmt *.go
