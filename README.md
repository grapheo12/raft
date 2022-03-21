# Raft Implementation 

## Raft Implementation done for Distributed Systems course

Authors:
1. Shubham Mishra (18CS10066)
2. Rounak Patra (18CS30048)
3. Archisman Pathak (18CS30050)

## Build Instructions 

In the project directory run:

`$ make`

The project is tested for Go v1.17.8

## Using CLI Key Value Store 
We have implemented a Key Value Store on top of the Raft Cluster. To use it : 
* Create a Raft Server  
    * Running the server binary and passing a config.ini file and the other parameters required.

     `$ ./bin/raft --help`

    * An example config.ini file can be found in `tests/perf_tests/config.ini`
    For full spec of the config file run:

    `$ go doc internal/config`

*  Read and write to the key-value store by executing the client binary with appropriate parameters  

    `$ ./bin/client --help`

## Run Tests

We implemented some unit tests and some performance tests

To run the unit tests, run:

`$ go test raft/tests/unit_tests`

In the performance test, we spawn 5 nodes
then bombard these nodes with client requests (both read and write) with different request rates
while manually pausing and resuming upto 2 nodes randomly.
Under these conditions, we measure the latency and throughput of the system.

To run this performance test, run:

`$ make perf_test`


## Package docs

To view the documentation for any package, run:

`$ go doc <package path>`

For example, to view the doc for the network package, run:

`$ go doc pkg/network`
