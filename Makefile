PROJECT='sharify-desktop'
build:
	@echo "Building ${PROJECT}"
	CGO_ENABLED=1 GO111MODULE=on go build -ldflags="-s -w" -o bin/${NAME}-dev .

run:
	./bin/${NAME}-${VERSION}