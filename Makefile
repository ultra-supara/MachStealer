.PHONY: all
all: test build

.PHONY: test
test:
	go mod tidy
	go test -v -race ./...

.PHONY: build
build:
	# xgo makes error when Terminal has multiple gopath, so define one GOPATH here
	# Only build for Apple Silicon Macs (arm64) - Intel Macs (amd64) are intentionally excluded
	GOPATH=${HOME}/go/ ${HOME}/go/bin/xgo -dest bin -out MachStealer -ldflags "-s -w" -targets darwin/arm64 ./
