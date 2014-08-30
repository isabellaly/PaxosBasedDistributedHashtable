package tests

import "testing"
import "strconv"
import "time"
import "fmt"
import "server"
import "math/rand"

type FakeAgent struct {
	agentID int
	address []string
	servers []server.Server
}

func MakeFakeAgent(servers []server.Server) *FakeAgent {
	fa := new(FakeAgent)
	fa.servers = servers
	fa.agentID = rand.Int()
	return fa
}

func (fa *FakeAgent) Put(key string, value string) {
	args := &server.PutArgs{fa.agentID, time.Now().UnixNano(), key, value}
	for {
		index := rand.Int() % len(fa.servers)
		reply := &server.PutReply{}
		if fa.servers[index] != nil {
			err := fa.servers[index].Put(args, reply)
			if err != nil {
				fmt.Println("err:", err)
			}
			if reply.OK {
				return 
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (fa *FakeAgent) Get(key string) string {
	args := &server.GetArgs{fa.agentID, time.Now().UnixNano(), key}
	for {
		index := rand.Int() % len(fa.servers)
		reply := &server.GetReply{}
		if fa.servers[index] != nil {
			err := fa.servers[index].Get(args, reply)
			if err != nil {
				fmt.Println("err:", err)
			}
			if reply.OK {
				return reply.Value
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (fa *FakeAgent) GetFrom(key string, serverID int) string {
	args := &server.GetArgs{fa.agentID, time.Now().UnixNano(), key}
	reply := &server.GetReply{}
	fa.servers[serverID].Get(args, reply)
	return reply.Value
}

func CreateAddress(port int) string {
	s := "localhost:"
	s += strconv.Itoa(port + 10000)
	return s
}

func Close(servers []server.Server) {
	for i := 0; i < len(servers); i++ {
		if servers[i] != nil {
			servers[i].Close()
		}
	}
}
func (fa *FakeAgent) Assess(t *testing.T, key string, expectValue string) {
	actualValue := fa.Get(key)
	if actualValue != expectValue {
		t.Fatalf("key: %v -> actual: %v but expected: %v", actualValue, actualValue, expectValue)
	}
}

func TestBasic(t *testing.T) {

	const serverNum = 3

	fmt.Printf("Basic Test: 3 servers put/get ...\n")

	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}
	var err error
	for i := 0; i < serverNum; i++ {
		servers[i], err = server.NewServer(address, i, false, false)
		if servers[i] == nil {
			fmt.Println("err:", err)
		}
	}

	ag := MakeFakeAgent(servers)
	var agent [serverNum]*FakeAgent
	for i := 0; i < serverNum; i++ {
		agent[i] = MakeFakeAgent([]server.Server{servers[i]})
	}

	ag.Put("key", "value1")
	ag.Assess(t, "key", "value1")

	agent[1].Put("key", "value2")
	agent[2].Assess(t, "key", "value2")

	agent[1].Assess(t, "key", "value2")
	ag.Assess(t, "key", "value2")

	fmt.Printf("  ... Passed\n")
	time.Sleep(1 * time.Second)

}

func TestOrder(t *testing.T) {

	const serverNum = 10
	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	fmt.Printf("Order Test: 10 servers put/get Assess order ...\n")
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}

	for i := 0; i < serverNum; i++ {
		servers[i], _ = server.NewServer(address, i, false, false)
	}

	fa := MakeFakeAgent(servers)
	var agent [serverNum]*FakeAgent
	for i := 0; i < serverNum; i++ {
		agent[i] = MakeFakeAgent(servers)
	}

	for i := 0; i < serverNum; i++ {
		agent[i].Put("key1", strconv.Itoa(i))
	}

	for i := 0; i < serverNum; i++ {
		agent[i].Put("key2", strconv.Itoa(i))
	}

	fa.Assess(t, "key1", strconv.Itoa(serverNum-1))
	fa.Assess(t, "key2", strconv.Itoa(serverNum-1))

	fmt.Printf("  ... Passed\n")

	time.Sleep(1 * time.Second)

}

func TestConcurrent(t *testing.T) {

	const serverNum = 5
	fmt.Printf("Concurrent Test: send requests at same time between 5 servers ...\n")

	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}
	for i := 0; i < serverNum; i++ {
		servers[i], _ = server.NewServer(address, i, false, false)
	}

	var agent [serverNum]*FakeAgent
	for i := 0; i < serverNum; i++ {
		agent[i] = MakeFakeAgent(servers)
	}

	var finish [serverNum]chan int
	for i := 0; i < serverNum; i++ {
		finish[i] = make(chan int)
		go func(me int) {
			defer func() { finish[me] <- 0 }()
			agent[me].Put("key", strconv.Itoa(rand.Int()))
		}(i)
	}
	for i := 0; i < serverNum; i++ {
		<-finish[i]
	}
	var value [serverNum]string
	for i := 0; i < serverNum; i++ {
		value[i] = agent[i].Get("key")
		if value[i] != value[0] {
			t.Fatalf("Value not match with input!")
		}
	}
	fmt.Printf("  ... Passed\n")

	time.Sleep(2 * time.Second)
}

func TestRobust(t *testing.T) {

	const serverNum = 3
	const requestEachServer = 300
	fmt.Printf("Robust Test 1:  %d servers %d request \n", serverNum, serverNum*requestEachServer)

	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}
	for i := 0; i < serverNum; i++ {
		servers[i], _ = server.NewServer(address, i, false, false)
	}

	ag := MakeFakeAgent(servers)

	for iters := 0; iters < requestEachServer; iters++ {
		for i := 0; i < serverNum; i++ {
			ag.Put("key", "value")
		}
	}

	for i := 0; i < serverNum; i++ {
		ag.GetFrom("key", i)
		if servers[i].StorageSize() != requestEachServer*serverNum+i+1 {
			t.Fatalf("wrong request numbers for server", i)
		}
	}
	fmt.Printf("  ... Passed\n")
}
