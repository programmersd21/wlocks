.PHONY: build install clean test

build:
	go build -o wlocks ./cmd/wlocks

install: build
	sudo mv wlocks /usr/local/bin/

clean:
	rm -f wlocks

test:
	go test -v ./...

run: build
	./wlocks

deps:
	go mod download
	go mod tidy
