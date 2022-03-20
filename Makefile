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

.PHONY: perf_test
perf_test: bin/raft bin/client tests/perf_tests/perf_test.py tests/perf_tests/config.ini
	python3 tests/perf_tests/perf_test.py
