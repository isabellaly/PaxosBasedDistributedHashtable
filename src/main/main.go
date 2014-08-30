package main 
	
import (
	"net/http"
	"agent"
	"server"
	"fmt"
	"strconv"
	"os"
)

func CreateAddress(port int) string {
	s := "localhost:"+strconv.Itoa(port + 10000)
	return s
}

func main() {
	serverNum, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Wrong server number")
		return
	}
	
	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}

	for i := 0; i < serverNum; i++ {
		servers[i], err = server.NewServer(address, i, false, false)
		if servers[i] == nil {
			fmt.Println("err:", err)
		}
	}

	a1, err := agent.NewAgent(address, 11, "10007")
	fmt.Println("a1 started")
	if err!=nil {
		fmt.Println(err)
		return
	}
	a2, err := agent.NewAgent(address, 22, "10008")
	fmt.Println("a2 started")
	if err!=nil {
		fmt.Println(err)
		return
	}
	a3, err := agent.NewAgent(address, 33, "10009")
	fmt.Println("a3 started")
	if err!=nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/Kiku/Get/", a1.GetHandler)
	http.HandleFunc("/Kiku/Put/", a1.PutHandler)
	go func(){
		fmt.Println("localhost:"+a1.Port)
		http.ListenAndServe(":10007", nil)
		select{}
	 }()

	go func(){
		fmt.Println("localhost:"+a2.Port)
		http.ListenAndServe(":10008", nil)
	}()

	fmt.Println("localhost:"+a3.Port)
	http.ListenAndServe(":10009", nil)
	select{}
}

