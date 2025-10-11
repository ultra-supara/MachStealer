.PHONY: all
all: test build

.PHONY: test
test:
	go mod tidy
	go test -v -race ./...

.PHONY: build
build:
	# xgo makes error when Terminal has multiple gopath, so define one GOPATH here
	GOPATH=${HOME}/go/ ${HOME}/go/bin/xgo -dest bin -out hcd -ldflags "-s -w" -targets darwin/amd64,darwin/arm64 ./
