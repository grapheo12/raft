package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"raft/pkg/server"

	"github.com/akamensky/argparse"
)

func readReq(addr string) {
	resp, err := http.Get("http://" + addr + "/read")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var msgs server.ReadResp
	err = json.Unmarshal(body, &msgs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Result:", msgs.Msg)
}

func writeReq(addr, msg string) {
	req := server.WriteReq{
		Content: msg,
	}
	data, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	r := bytes.NewReader(data)
	resp, err := http.Post("http://"+addr+"/write", "application/json", r)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if resp.StatusCode == http.StatusBadGateway {
		body := make(map[string]string)
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		err = json.Unmarshal(data, &body)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		newAddr, ok := body["leader"]
		if !ok {
			fmt.Println("Malformed response")
			return
		}

		writeReq(newAddr, msg)
	}

	if resp.StatusCode == http.StatusCreated {
		fmt.Println("Write by", addr)
	}
}

func main() {
	parser := argparse.NewParser("", "Raft Client")
	reader := parser.NewCommand("read", "Given an address, reads its log")
	writer := parser.NewCommand("write", "Writes a value to leader")
	addr := parser.String("a", "addr", &argparse.Options{
		Required: true,
		Help:     "IP:Port to a node",
	})
	msg := writer.String("m", "msg", &argparse.Options{
		Required: true,
		Help:     "Message to send",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(parser.Usage(err))
		return
	}

	switch {
	case reader.Happened():
		readReq(*addr)
	case writer.Happened():
		writeReq(*addr, *msg)
	}

}
