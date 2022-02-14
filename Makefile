GOFILES = $(shell find . -name "*.go")

all: bin/raft bin/client

bin/raft: cmd/server/main.go $(GOFILES)
	mkdir -p bin
	go build -o bin/raft cmd/server/main.go

bin/client: cmd/client/main.go $(GOFILES)
	mkdir -p bin
	go build -o bin/client cmd/client/main.go

.PHONY: clean
clean:
	rm -rf bin
	rm -f *.log

.PHONY: protocompile
protocompile:
	cd pkg/rpc && protoc --gofast_out=. network.proto && cd ../../
	cd pkg/rpc && protoc --gofast_out=. raft_msg.proto && cd ../../
