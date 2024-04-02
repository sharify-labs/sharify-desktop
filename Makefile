PROJECT='sharify-desktop'
build:
	@echo "Building ${PROJECT}"
	CGO_ENABLED=1 GO111MODULE=on go build -ldflags="-s -w -X main.Version=dev" -o bin/${PROJECT}-dev .

run:
	./bin/${PROJECT}-dev