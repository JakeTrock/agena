# Agena Makefile

.PHONY: build test clean smoke-test docker-build docker-test docker-smoke-test

# Local Build Commands
build:
	go build -o bin/agena ./src

test:
	go test -v ./src/...

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
