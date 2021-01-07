package agent

import (
	"github.com/astaxie/beego/logs"
	"net"

	"proto_trace/Lyingdragon_Protocol"
	. "proto_trace/session"
	"proto_trace/trace"
	"proto_trace/util"
)

type AgentData struct {
	loginname string
	key       string
}

func Run(src, dst string) {
	srcAddr, err := net.ResolveTCPAddr("tcp4", src)
	if err != nil {
		logs.Error("ResolveTCPAddr", err)
		return
	}
	dstAddr, err := net.ResolveTCPAddr("tcp4", dst)
	if err != nil {
		logs.Error("ResolveTCPAddr", err)
		return
	}
	listen, err := net.ListenTCP("tcp", srcAddr)
	if err != nil {
		logs.Error("ListenTCP", err)
		return
	}
	for {
		srcConn, err := listen.AcceptTCP()
		if err != nil {
			logs.Error("AcceptTCP", err)
			continue
		}
		logs.Info("Connected from", srcConn.RemoteAddr())
		data := make([]byte, 4)
		srcConn.Read(data)

		dstConn, err := net.DialTCP("tcp", nil, dstAddr)
		if err != nil {
			logs.Error("DialTCP", err)
			srcConn.Close()
			continue
		}
		logs.Info("Connected to", dstConn.RemoteAddr())
		dstConn.Write(data)

		uuid := util.GetUUID()

		sess := NewSession(srcConn, uuid)
		sess.SetReadCallback(onSrcMessage)
		sess.SetDisconnectCallback(onSrcDisconnect)
		sess.Data = &AgentData{}
		g_srcMan.AddSession(sess)
		sess.Run()

		sess = NewSession(dstConn, uuid)
		sess.SetReadCallback(onDstMessage)
		sess.SetDisconnectCallback(onDstDisconnect)
		sess.Data = &AgentData{}
		g_dstMan.AddSession(sess)
		sess.Run()
	}
}

func onSrcMessage(sess *Session, buff []byte) {
	sessDst := g_dstMan.GetSession(sess.Id())
	if sessDst == nil {
		sess.Close()
		return
	}
	sessDst.Send(buff)

	buff, err := util.AesDecrypt(buff, []byte(sess.Data.(*AgentData).key))
	if err != nil {
		logs.Error(err, "key", sess.Data.(*AgentData).key)
		return
	}
	t, msg := util.Buff2ProtoMsg(buff[20:])
	if msg == nil {
		logs.Error("Buff2ProtoMsg error", t)
		return
	}
	if t == int32(Lyingdragon_Protocol.ProtocolType_C2S_CONNECT_TO_SERVER) {
		req := msg.(*Lyingdragon_Protocol.C2S_ConnectToServer)
		sess.Data.(*AgentData).loginname = req.GetLoginname()
		sessDst.Data.(*AgentData).loginname = sess.Data.(*AgentData).loginname
	}
	trace.PushLog(trace.NewLog(sess.Data.(*AgentData).loginname, true, t, msg))
}

func onDstMessage(sess *Session, buff []byte) {
	sessSrc := g_srcMan.GetSession(sess.Id())
	if sessSrc == nil {
		sess.Close()
		return
	}
	sessSrc.Send(buff)

	t, msg := util.Buff2ProtoMsg(buff)
	if msg == nil {
		logs.Error("Buff2ProtoMsg error", t)
		return
	}
	if t == int32(Lyingdragon_Protocol.ProtocolType_S2C_GATE_STATE) {
		req := msg.(*Lyingdragon_Protocol.S2C_GateState)
		sess.Data.(*AgentData).key = GetSessionKey(req.GetKey())
		sessSrc.Data.(*AgentData).key = sess.Data.(*AgentData).key
	}
	trace.PushLog(trace.NewLog(sess.Data.(*AgentData).loginname, false, t, msg))
}

func onSrcDisconnect(sess *Session) {
	g_srcMan.RemoveSession(sess)
}

func onDstDisconnect(sess *Session) {
	g_dstMan.RemoveSession(sess)
}

func GetSessionKey(key int64) string {
	EncryptDict := "P%2BViyZLtO^gRT2Huxqx#5VygbflvZx$8mFpX61VWvd;ivPu~XjL`CD7FrIe8=0"
	ret := make([]byte, 64)
	index := 0
	for i := 63; i >= 0; i-- {
		if key&(int64(1)<<i) > 0 {
			if i < len(EncryptDict) {
				ret[index] = EncryptDict[len(EncryptDict)-1-i]
				index++
			}
		}
	}
	for index < 32 {
		ret[index] = byte('a')
		index++
	}
	return string(ret[:32])
}

func init() {
	g_srcMan = NewSessionManager()
	g_dstMan = NewSessionManager()
}

var g_srcMan *SessionManager
var g_dstMan *SessionManager
