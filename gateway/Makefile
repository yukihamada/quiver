.PHONY: build test clean run

build:
	go build -o bin/gateway cmd/gateway/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

run: build
	./bin/gateway