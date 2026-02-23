# Agena Makefile

.PHONY: build test clean install smoke-test docker-build docker-test docker-smoke-test

PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install

# Local Build Commands
build:
	go build -o bin/agena ./src

test:
	go test -v ./src/...

install: build
	$(INSTALL) -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 755 bin/agena $(DESTDIR)$(BINDIR)/agena

clean:
	rm -rf bin/
	rm -f test-environment/agena/*/*.log
	rm -f test-environment/.fixed-*

smoke-test: build
	cd test-environment && ./test-smoke.sh --non-interactive

# Docker Build Commands
docker-build:
	docker build -t agena .

docker-test: docker-build
	docker run --rm --entrypoint go agena test -v ./...

docker-smoke-test: docker-build
	docker run --rm --entrypoint bash agena -c "cd test-environment && ./test-smoke.sh --non-interactive"
