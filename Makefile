export GO111MODULE=on

build:
	go build -o build/emd ./cmd/daemon
	go build -o build/emcli ./cmd/cli

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

clean:
	rm -rf ./build ./data ./config
