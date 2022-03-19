// Package config consists of config parser for the raft implementation.
//
// Configs can be sent in both an ini file and in command line arguments.
// In case a particular config is sent in both, command line takes precedence over ini file.
//
// Full spec of ini file:
//
//
//     [network]
//     network-members = <comma separated list of ip:port for Raft messages, containing self>
//     id = <Index of self in the above list>
//     network-ip = <Ip to listen to for Raft messages, may be different from above>
//     network-port = <Port to listen to for Raft messages, may be different from above>
//
//     [raft]
//     min-election-timeout = <time in ms>
//     max-election-timeout = <time in ms>
//
//     [client]
//     client-ip = <Ip to listen to for client messages>
//     client-port = <Port to listen to for client messages>
//     client-members = <comma separated list of ip:port for client messages, containing self>
//
//     [log]
//     log-level = <Log Level to use, check internal/lo>
//
// All options are available as command line arguments also.
// To specify the ini file to use, use --config argument.
package config
