PROJECT='sharify-desktop'

.PHONY: all audit build clean lint run tidy

all: audit lint build

audit:
	go mod verify
	go vet ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

build: clean
	CGO_ENABLED=1 GO111MODULE=on go build -ldflags="-s -w -X main.Version=dev" -o bin/${PROJECT}-dev .

clean:
	rm -rf ./bin

lint: tidy
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.57.2 run ./...

run:
	./bin/${PROJECT}-dev

tidy:
	go mod tidy -v
	go run mvdan.cc/gofumpt@latest -w -l .