package trace

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"

	"proto_trace/parse"
)

const (
	MAX_LOG_NUM = 2000
	LOG_POP_NUM = 10
)

type Log struct {
	Ts        string `json:"ts"`
	Loginname string `json:"loginname"`
	Dir       string `json:"dir"`
	Id        int32  `json:"id"`
	Msg       string `json:"msg"`
	Detail    string `json:"detail"`
}

func NewLog(loginname string, dir bool, id int32, msg proto.Message) *Log {
	log := &Log{
		Id:        id,
		Loginname: loginname,
	}
	ts := time.Now().UnixNano()
	log.Ts = fmt.Sprintf("%s.%06d", time.Unix(ts/1e9, 0).Format("2006-01-02 15:04:05"), ts%1e9/1e3)
	if dir {
		log.Dir = "send"
	} else {
		log.Dir = "recv"
	}
	log.Msg = parse.GetMessageNameById(id)
	log.Detail = msg.String()
	return log
}

func PushLog(log *Log) {
	g_lock.Lock()
	defer g_lock.Unlock()
	g_logs = append(g_logs, log)
	if len(g_logs) > MAX_LOG_NUM {
		g_logs = g_logs[len(g_logs)-MAX_LOG_NUM:]
	}
}

func PopLog() []*Log {
	g_lock.Lock()
	defer g_lock.Unlock()
	var ret []*Log
	if len(g_logs) <= LOG_POP_NUM {
		ret = g_logs
		g_logs = make([]*Log, 0, LOG_POP_NUM)
	} else {
		ret = make([]*Log, LOG_POP_NUM)
		copy(ret, g_logs[:LOG_POP_NUM])
		g_logs = g_logs[LOG_POP_NUM:]
	}
	return ret
}

var g_lock sync.Mutex
var g_logs []*Log
