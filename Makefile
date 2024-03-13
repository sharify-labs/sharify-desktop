VERSION=$(shell cat VERSION)
NAME='zephyr-desktop'
build:
	@echo "Building version ${VERSION}"
	CGO_ENABLED=1 GO111MODULE=on go build -ldflags="-s -w" -o bin/${NAME}-${VERSION} .

run:
	./bin/${NAME}-${VERSION}