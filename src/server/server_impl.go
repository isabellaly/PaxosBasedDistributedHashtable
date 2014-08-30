package server

import (
	"encoding/gob"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"paxos"
	"strconv"
	"strings"
	"sync"
	"time"
)

type server struct {
	allHostPorts []string
	self         int
	rid          int
	p            paxos.Paxos
	storage      map[string]string
	ridLock      *sync.Mutex
	storageLock  *sync.Mutex
	closeLock    *sync.Mutex
	listener     net.Listener
	closed       bool
	needFile     bool
	fileName     string
}

func NewServer(allHostPorts []string, self int, isDebug bool, needFile bool) (Server, error) {
	gob.Register(Request{})
	s := &server{
		allHostPorts: allHostPorts,
		self:         self,
		rid:          0,
		storage:      make(map[string]string),
		ridLock:      new(sync.Mutex),
		storageLock:  new(sync.Mutex),
		closeLock:    new(sync.Mutex),
		closed:       false,
		needFile:     needFile,
		fileName:     "../logs/log_" + allHostPorts[self],
	}
	if needFile {
		if _, err := os.Stat(s.fileName); os.IsNotExist(err) {
			s.writeFile(os.O_CREATE|os.O_TRUNC|os.O_RDWR, "Server: "+allHostPorts[self]+"\n")
		} else {
			s.recovery()
		}
	}
	newRpc := rpc.NewServer()
	p := paxos.NewPaxos(allHostPorts, self, newRpc, isDebug)
	s.p = p
	err := newRpc.RegisterName("Server", Wrap(s))
	if err != nil {
		return nil, err
	}
	s.listener, err = net.Listen("tcp", allHostPorts[self])
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			conn, err := s.listener.Accept()
			s.closeLock.Lock()
			if err == nil && !s.closed {
				go newRpc.ServeConn(conn)
			} else if err == nil {
				conn.Close()
			} else if s.closed {
				break
			}
			s.closeLock.Unlock()
		}
	}()
	return s, nil
}

func (s *server) Get(args *GetArgs, reply *GetReply) error {
	s.ridLock.Lock()
	defer s.ridLock.Unlock()
	r := Request{}
	r.AgentID = args.AgentID
	r.RequestID = args.RequestID
	r.Name = "Get"
	r.Key = args.Key
	reply.OK = true

	for {
		// start a new round of paxos
		s.p.StartPaxos(s.rid, r)
		var new_r Request
		// get log
		for {
			commit, log_r := s.p.GetLog(s.rid)
			if commit {
				new_r = log_r.(Request)
				break
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
		// if get another r, means this round has been performed by other paxos node
		// and it should catch up with other
		if s.needFile {
			s.writeFile(os.O_APPEND|os.O_RDWR, s.genText(new_r))
		}
		if new_r.Name == "Put" {
			s.storageLock.Lock()
			s.storage[new_r.Key] = new_r.Value
			s.storageLock.Unlock()
		}
		if new_r.RequestID == r.RequestID && new_r.AgentID == r.AgentID {
			s.storageLock.Lock()
			if v, ok := s.storage[r.Key]; ok {
				s.storageLock.Unlock()
				reply.AgentID = r.AgentID
				reply.RequestID = r.RequestID
				reply.Value = v
				break
			} else {
				s.storageLock.Unlock()
				reply.AgentID = r.AgentID
				reply.RequestID = r.RequestID
				reply.OK = false
				break
			}
		}
		s.rid++
	}
	s.p.CommitFinished(s.rid)
	s.rid++
	if reply.OK {
		return nil
	} else {
		return errors.New("Could not find the Key in storage")
	}

}

func (s *server) Put(args *PutArgs, reply *PutReply) error {

	s.ridLock.Lock()
	defer s.ridLock.Unlock()
	r := Request{}
	r.AgentID = args.AgentID
	r.RequestID = args.RequestID
	r.Name = "Put"
	r.Key = args.Key
	r.Value = args.Value
	reply.OK = true

	

	for {
		// start a new round of paxos
		s.p.StartPaxos(s.rid, r)
		var new_r Request
		// get log
		for {
			commit, log_r := s.p.GetLog(s.rid)
			if commit {
				new_r = log_r.(Request)
				break
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}
		// if get another r, means this round has been performed by other paxos node
		// and it should catch up with other
		if s.needFile {
			s.writeFile(os.O_APPEND|os.O_RDWR, s.genText(new_r))
		}
		if new_r.Name == "Put" {
			s.storageLock.Lock()
			s.storage[new_r.Key] = new_r.Value
			s.storageLock.Unlock()
		}
		if new_r.AgentID == r.AgentID && new_r.RequestID == r.RequestID {
			reply.AgentID = r.AgentID
			reply.RequestID = r.RequestID
			break
		}
		s.rid++
	}
	s.p.CommitFinished(s.rid)
	s.rid++

	return nil

}

func (s *server) genText(new_r Request) string {
	var text = new_r.Name + "::" + new_r.Key + "::" + new_r.Value + "::" + strconv.Itoa(int(new_r.RequestID)) + "::" + strconv.Itoa(new_r.AgentID) + "::" + strconv.Itoa(s.rid) + "\n"
	return text
}
func (s *server) writeFile(flag int, text string) {
	f, _ := os.OpenFile(s.fileName, flag, 0666)
	defer f.Close()
	io.WriteString(f, text)
}

func (s *server) recovery() {
	fileBytes, err := ioutil.ReadFile(s.fileName)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(fileBytes), "\n") {
		e := strings.Split(string(line), "::")
		if e[0] == "Put" {
			s.storage[e[1]] = e[2]
			s.rid++
		}else if e[0] == "Get"{
			s.rid++
		}
	}

}

func (s *server) Close() {
	s.listener.Close()
	s.p.Close()
	s.closeLock.Lock()
	s.closed = true
	s.closeLock.Unlock()
}

func (s *server) StorageSize() int {
	return s.rid
}
