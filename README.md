# Raft Implementation 
## Raft Implementation done for Distributed Systems course

## Build Instructions 
In the project directory run   
`$ make`

## Using CLI Key Value Store 
We have implemented a Key Value Store on top of the Raft Cluster. To use it : 
* Create a Raft Server  
    * Running the server binary and passing a config.ini file and the other parameters required.
     `$ ./bin/raft --help` 
    * An example config.ini file can be found in `tests/perf_tests/config.ini`

*  Read and write to the key-value store by executing the client binary with appropriate parameters  
    `$ ./bin/client --help`

## Run Tests
We implemented some unit tests and some performance tests
`$ go test raft/tests/unit_tests`