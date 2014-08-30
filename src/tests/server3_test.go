package tests

import "testing"
import "server"
import "fmt"

func TestDrop(t *testing.T) {

	const serverNum = 3
	const requestEachServer = 5
	fmt.Printf("Drop Package Test : 50 percent lost package\n")

	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}
	for i := 0; i < serverNum; i++ {
		servers[i], _ = server.NewServer(address, i, true, false)
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

func TestStress(t *testing.T) {

	const serverNum = 10
	const requestEachServer = 200
	fmt.Printf("Stress Test :  %d servers %d request \n", serverNum, serverNum*requestEachServer)

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
