.PHONY: default

default: swagger-update mock run

init:
	brew install vektra/tap/mockery
	brew upgrade mockery
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0
	go install github.com/swaggo/swag/cmd/swag@v1.8.2
	export PATH=$PATH:$(go env GOPATH)/bin/swag
	export GOPRIVATE=*.trendyol.com

clean:
	rm -rf ./build
	rm -rf mocks

linter: mock
	true || golangci-lint run ./...

test: mock
	go test -v -timeout 30s -coverprofile=coverage.out -cover ./...
	go tool cover -func=coverage.out

build: clean swagger-update mock test linter
	CGO_ENABLED=0 go build -ldflags="-w -s"

swagger-update:
	swag init

mock:
	go install github.com/vektra/mockery/v2@v2.14.0
	go generate ./...

run:
	go run main.go

mutation-test: mock
	go get -t -v github.com/zimmski/go-mutesting/...
	true || go-mutesting ./application/...  ./pkg/...
