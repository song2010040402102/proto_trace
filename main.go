package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/astaxie/beego/logs"

	"proto_trace/agent"
	"proto_trace/config"
	"proto_trace/parse"
	"proto_trace/trace"
)

func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, _ := json.Marshal(trace.PopLog())
		fmt.Fprint(w, string(data))
	} else {
		fmt.Fprint(w, r.Method, "not support!")
	}
}

func main() {
	logs.Notice("server start...")
	parse.ParseProto(config.Get().Protocol)

	http.Handle("/", http.FileServer(http.Dir(config.Get().Web)))
	http.HandleFunc("/log", logHandler)
	go http.ListenAndServe(config.Get().Http, nil)

	go agent.Run(config.Get().Agent.Src, config.Get().Agent.Dst)

	var sig os.Signal
	c := make(chan os.Signal, 1)
	for {
		signal.Notify(c)
		sig = <-c
		if sig != syscall.SIGPIPE && sig != syscall.SIGCHLD && sig != syscall.SIGURG {
			break
		}
	}
	logs.Notice("server terminate with sig:", sig)
}
