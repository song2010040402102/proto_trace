package session

import (
	"net"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	MAX_WRITE_CACHE int = 32
	READ_BUFF_SIZE  int = 1024
	WRITE_BUFF_SIZE int = 1024
	READ_TIMEOUT    int = 120
	WRITE_TIMEOUT   int = 60
)

type Session struct {
	conn               *net.TCPConn
	id                 uint64
	readCallback       func(*Session, []byte)
	disconnectCallback func(*Session)
	writeCache         chan []byte
	Data               interface{}
}

func NewSession(conn *net.TCPConn, id uint64) *Session {
	conn.SetReadBuffer(READ_BUFF_SIZE)
	conn.SetWriteBuffer(WRITE_BUFF_SIZE)
	conn.SetNoDelay(false)
	return &Session{
		conn:       conn,
		id:         id,
		writeCache: make(chan []byte, MAX_WRITE_CACHE),
	}
}

func (cs *Session) Id() uint64 {
	return cs.id
}

func (cs *Session) SetReadCallback(cb func(*Session, []byte)) {
	cs.readCallback = cb
}

func (cs *Session) SetDisconnectCallback(cb func(*Session)) {
	cs.disconnectCallback = cb
}

func (cs *Session) Run() {
	go cs.read()
	go cs.write()
}

func (cs *Session) Send(buff []byte) {
	cs.writeCache <- buff
}

func (cs *Session) Close() {
	cs.conn.Close()
	if cs.disconnectCallback != nil {
		cs.disconnectCallback(cs)
	}
}

func (cs *Session) read() {
	buff := make([]byte, 0, READ_BUFF_SIZE)
	size := uint32(0)
	for {
		var data []byte
		b := make([]byte, READ_BUFF_SIZE)
		cs.conn.SetReadDeadline(time.Now().Add(time.Duration(READ_TIMEOUT) * time.Second))
		n, err := cs.conn.Read(b)
		if err != nil {
			logs.Error("session read", n, err, cs.conn.RemoteAddr())
			cs.Close()
			return
		}
		buff = append(buff, b[:n]...)
		if len(buff) < 4 {
			continue
		}
		if size == 0 {
			size = uint32(buff[0]) | uint32(buff[1])<<8 | uint32(buff[2])<<16 | uint32(buff[3])<<24
		}
		for len(buff) >= int(size)+4 {
			data = make([]byte, size-4)
			copy(data, buff[4:size])
			buff = buff[size:]
			size = uint32(buff[0]) | uint32(buff[1])<<8 | uint32(buff[2])<<16 | uint32(buff[3])<<24
			if cs.readCallback != nil {
				cs.readCallback(cs, data)
			}
		}
		if len(buff) == int(size) {
			data = make([]byte, size-4)
			copy(data, buff[4:size])
			buff = buff[:0]
			size = 0
			if cs.readCallback != nil {
				cs.readCallback(cs, data)
			}
		}
	}
}

func (cs *Session) write() {
	for {
		buff := make([]byte, 0, WRITE_BUFF_SIZE)
		for {
			b := <-cs.writeCache
			if len(b) == 0 {
				continue
			}
			size := uint32(len(b)) + 4
			buff = append(buff, byte(size&0xff))
			buff = append(buff, byte(size>>8&0xff))
			buff = append(buff, byte(size>>16&0xff))
			buff = append(buff, byte(size>>24&0xff))
			buff = append(buff, b...)
			if len(cs.writeCache) == 0 || len(buff) >= WRITE_BUFF_SIZE {
				break
			}
		}
		for i := 0; i < len(buff); {
			j := i + WRITE_BUFF_SIZE
			if j > len(buff) {
				j = len(buff)
			}
			cs.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
			n, err := cs.conn.Write(buff[i:j])
			if err != nil {
				logs.Error("session write", n, err)
				cs.Close()
				return
			}
			i += n
		}
	}
}

type SessionManager struct {
	lock       sync.RWMutex
	mapSession map[uint64]*Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		mapSession: make(map[uint64]*Session),
	}
}

func (sm *SessionManager) GetSession(id uint64) *Session {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return sm.mapSession[id]
}

func (sm *SessionManager) AddSession(sess *Session) {
	if sess == nil {
		return
	}
	sm.lock.Lock()
	defer sm.lock.Unlock()
	if _, ok := sm.mapSession[sess.id]; ok {
		return
	}
	sm.mapSession[sess.id] = sess
}

func (sm *SessionManager) RemoveSession(sess *Session) {
	if sess == nil {
		return
	}
	sm.lock.Lock()
	defer sm.lock.Unlock()
	delete(sm.mapSession, sess.id)
}
