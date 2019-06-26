export GO111MODULE=on

build:
	go build -o build/emd ./cmd/daemon
	go build -o build/emcli ./cmd/cli

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

clean:
	rm -rf ./build ./data ./config
