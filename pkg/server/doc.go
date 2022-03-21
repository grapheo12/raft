// server package provides an HTTP server for client requests.
//
// Defined methods:
// - GET /ping : For heartbeat to client
// - GET /read : To return the committed log of the node
// - POST /write : To append a new entry to the log
package server
