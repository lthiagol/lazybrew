.PHONY: build run test test-integration lint fmt clean cover

build:
	go build -o bin/lazybrew ./cmd/lazybrew

run:
	go run ./cmd/lazybrew

test:
	go test ./... -v -race -count=1

test-integration:
	go test ./... -v -race -count=1 -tags=integration

lint:
	go vet ./...

fmt:
	gofmt -s -w .

clean:
	rm -rf bin/

cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out
