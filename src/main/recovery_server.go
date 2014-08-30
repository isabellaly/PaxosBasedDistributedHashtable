package main 

import (
	"fmt"
	"server"
	"os"
	"strconv"
)

func CreateAddress(port int) string {
	s := "localhost:"+strconv.Itoa(port + 11000)
	return s
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

	serverID, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Wrong serverID")
		return
	}

	_, err = server.NewServer(address, serverID, false, true)
	if err!=nil {
		fmt.Println("err: "+err.Error())
	}
	select {}
}
