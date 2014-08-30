package main 

import (
	"fmt"
	"time"
	"server"
	"os"
	"strconv"
	"net/rpc"
)

func CreateAddress(port int) string {
	s := "localhost:"+strconv.Itoa(port + 11000)
	return s
}

func agentPut(servers []*rpc.Client, key string, value string, callServer int, agentID int){
	timeStamp := time.Now().UnixNano()

	putArgs := &server.PutArgs{}
	putArgs.AgentID = agentID
	putArgs.RequestID = timeStamp
	putArgs.Key = key
	putArgs.Value = value

	var putReply server.PutReply

	fmt.Println("Agent put key: "+key+" value: "+value+" to server: "+strconv.Itoa(callServer))
	error := servers[callServer].Call("Server.Put", putArgs, &putReply)

	if error == nil && putReply.OK {
		fmt.Println("OK")
	} else {
		fmt.Println("Agent put error: "+error.Error())
	}
	return
}

func agentGet(servers []*rpc.Client, key string, serverNum int, agentID int){
	timeStamp := time.Now().UnixNano()

	getArgs := &server.GetArgs{}
	getArgs.AgentID = 0
	getArgs.RequestID = timeStamp
	getArgs.Key = key

	var getReply server.GetReply

	for callServer := 0; callServer<serverNum; callServer++ {
		error := servers[callServer].Call("Server.Get", getArgs, &getReply)
		if error == nil && getReply.OK {
			fmt.Println("Agent get key: "+key+" from server: "+strconv.Itoa(callServer)+ " and get value: "+ getReply.Value)
		} else {
			fmt.Println("Agent get error: "+error.Error())
		}
	}
	return
}

func main() {
	serverNum, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Wrong serverNum")
		return
	}
	var address []string = make([]string, serverNum)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}

	var servers []*rpc.Client = make([]*rpc.Client, serverNum)

	for i := 0; i < serverNum; i++ {
		servers[i], err = rpc.Dial("tcp", address[i])
		if err != nil {
			fmt.Println(err)
			return
		} 
	}

	agentID := 0

	for i:=0; i<serverNum; i++ {
		agentPut(servers, strconv.Itoa(i), strconv.Itoa(i), i, agentID)
	}

	for i:=0; i<serverNum; i++ {
		agentGet(servers, strconv.Itoa(i), serverNum, agentID)
	}

	fmt.Println("Please stop one server and press enter")
	var input string
	fmt.Scanf("%s\n", &input)

	for i:=0; i<serverNum; i++ {
		agentPut(servers, strconv.Itoa(i+serverNum), strconv.Itoa(i+serverNum), i, agentID)
	}

	for i:=0; i<serverNum; i++ {
		agentGet(servers, strconv.Itoa(i+serverNum), serverNum, agentID)
	}

	fmt.Println("Please restart that server and press enter")
	
	fmt.Scanf("%s\n", &input)

	for i := 0; i < serverNum; i++ {
		servers[i], err = rpc.Dial("tcp", address[i])
		if err != nil {
			fmt.Println(err)
			return
		} 
	}

	for i:=0; i<serverNum; i++ {
		agentPut(servers, strconv.Itoa(i+serverNum*2), strconv.Itoa(i+serverNum*2), i, agentID)
	}

	for i:=0; i<serverNum; i++ {
		agentGet(servers, strconv.Itoa(i+serverNum*2), serverNum, agentID)
	}

	return 
}

	


