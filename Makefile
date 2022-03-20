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
	@echo "Read : Write = 3 : 1"
	
	
	@echo "Rate: 50 req/s" 
	python3 tests/perf_tests/perf_test.py 50
	@rm *.log

	@echo "Rate: 100 req/s" 
	python3 tests/perf_tests/perf_test.py 100
	@rm *.log

	@echo "Rate: 200 req/s" 
	python3 tests/perf_tests/perf_test.py 200
	@rm *.log

	@echo "Rate: 400 req/s" 
	python3 tests/perf_tests/perf_test.py 400
	@rm *.log

	@echo "Rate: 600 req/s" 
	python3 tests/perf_tests/perf_test.py 600
	@rm *.log

	@echo "Rate: 900 req/s" 
	python3 tests/perf_tests/perf_test.py 900
	@rm *.log

	@echo "Rate: 1200 req/s" 
	python3 tests/perf_tests/perf_test.py 1200
	@rm *.log

	@echo "Rate: 1800 req/s" 
	python3 tests/perf_tests/perf_test.py 1800
	@rm *.log

	@echo "Rate: 2000 req/s" 
	python3 tests/perf_tests/perf_test.py 2000
	@rm *.log