GOFILES = $(shell find . -name "*.go")

all: bin/raft

bin/raft: cmd/main.go $(GOFILES)
	mkdir -p bin
	go build -o bin/raft cmd/main.go

.PHONY: clean
clean:
	rm -rf bin
	rm -f *.log

.PHONY: protocompile
protocompile:
	cd pkg/rpc && protoc --gofast_out=. network.proto && cd ../../
	cd pkg/rpc && protoc --gofast_out=. raft_msg.proto && cd ../../
