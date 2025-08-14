.PHONY: build test clean run

build:
	go build -o bin/provider cmd/provider/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

run: build
	./bin/provider