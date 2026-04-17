.PHONY: build run test clean install

build:
	go build -o bin/oc-go-cc ./cmd/oc-go-cc

run:
	go run ./cmd/oc-go-cc

test:
	go test ./...

clean:
	rm -rf bin/

install:
	go install ./cmd/oc-go-cc
