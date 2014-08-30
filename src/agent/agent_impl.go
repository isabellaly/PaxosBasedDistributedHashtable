package agent

import (
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
	"server"
	"strings"
	"time"
)

const (
	TryConnect = 10
)

type agent struct {
	allHostPorts []string
	servers      []*rpc.Client
	agentID      int
	Port         string
}

func NewAgent(allHostPorts []string, agentID int, port string) (*agent, error) {
	var err error
	a := &agent{}
	a.allHostPorts = allHostPorts
	a.agentID = agentID
	a.Port = port
	a.servers = make([]*rpc.Client, len(allHostPorts))

	for i := 0; i < len(allHostPorts); i++ {
		for j := 0; j < TryConnect; j++ {
			a.servers[i], err = rpc.Dial("tcp", allHostPorts[i])
			if err != nil {
				fmt.Println(err)
			} else {
				break
			}
		}
		if err != nil {
			error := errors.New("Could not start all servers")
			fmt.Println(error)
			return nil, error
		}
	}
	return a, nil
}

func (a *agent) GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/Kiku/Get/"):]
	timeStamp := time.Now().UnixNano()
	callServer := timeStamp % int64(len(a.servers))

	getArgs := &server.GetArgs{}
	getArgs.AgentID = a.agentID
	getArgs.RequestID = timeStamp
	getArgs.Key = key

	var getReply server.GetReply
	error := a.servers[callServer].Call("Server.Get", getArgs, &getReply)

	if error == nil && getReply.OK {
		fmt.Fprint(w, getReply.Value)
	} else {
		fmt.Fprint(w, "Agent get error: " + error.Error())
	}
	return
}

func (a *agent) PutHandler(w http.ResponseWriter, r *http.Request) {
	remPartOfURL := r.URL.Path[len("/Kiku/Put/"):]
	kvpair := strings.Split(remPartOfURL, "&")
	key := kvpair[0]
	value := kvpair[1]
	timeStamp := time.Now().UnixNano()
	callServer := timeStamp % int64(len(a.servers))

	putArgs := &server.PutArgs{}
	putArgs.AgentID = a.agentID
	putArgs.RequestID = timeStamp
	putArgs.Key = key
	putArgs.Value = value

	var putReply server.PutReply
	error := a.servers[callServer].Call("Server.Put", putArgs, &putReply)

	if error == nil && putReply.OK {
		fmt.Fprint(w, "OK")
	} else {
		fmt.Fprint(w, "Agent put error: "+error.Error())
	}
	return
}
