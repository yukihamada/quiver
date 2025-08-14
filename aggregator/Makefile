.PHONY: build test clean run

build:
	go build -o bin/aggregator cmd/aggregator/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

run: build
	./bin/aggregator