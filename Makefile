.PHONY: build clean test install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o bin/repokill ./cmd/repokill

clean:
	rm -rf bin/

test:
	go test -v ./...

install: build
	install -Dm755 bin/repokill $(PREFIX)/usr/local/bin/repokill
