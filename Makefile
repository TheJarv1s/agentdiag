BINARY=agentdiag
VERSION=0.1.0

.PHONY: test vet build release clean

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -trimpath -o $(BINARY) ./cmd/agentdiag

release: clean
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/$(BINARY)-v$(VERSION)-windows-amd64.exe ./cmd/agentdiag
	GOOS=windows GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/$(BINARY)-v$(VERSION)-windows-arm64.exe ./cmd/agentdiag
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/$(BINARY)-v$(VERSION)-linux-amd64 ./cmd/agentdiag
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/$(BINARY)-v$(VERSION)-linux-arm64 ./cmd/agentdiag
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o dist/$(BINARY)-v$(VERSION)-darwin-amd64 ./cmd/agentdiag
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o dist/$(BINARY)-v$(VERSION)-darwin-arm64 ./cmd/agentdiag

clean:
	rm -rf dist $(BINARY) $(BINARY).exe
