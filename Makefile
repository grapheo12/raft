all: bin/raft

bin/raft: cmd/main.go pkg
	mkdir -p bin
	go build -o bin/raft cmd/main.go

.PHONY: clean
clean:
	rm -rf bin

.PHONY: protocompile
protocompile:
	cd pkg/rpc && protoc --gofast_out=. network.proto && cd ../../