package tests

import "testing"
import "time"
import "fmt"
import "server"

func TestDeadServer1(t *testing.T) {
	const serverNum = 5
	const deadNum = 2
	fmt.Printf("TestDeadServer Test 1:  test %d servers and %d of them is dead \n", serverNum, deadNum)

	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i)
	}
	for i := 0; i < serverNum-deadNum; i++ {
		servers[i], _ = server.NewServer(address, i, false, false)
	}

	var agent [serverNum]*FakeAgent
	for i := 0; i < serverNum-deadNum; i++ {
		agent[i] = MakeFakeAgent(servers)
	}

	for i := 0; i < serverNum-deadNum; i++ {
		agent[i].Put("key1", "value1")
		agent[i].Put("key2", "value2")
	}

	var va [serverNum]string
	for i := 0; i < serverNum-deadNum; i++ {
		va[i] = agent[i].Get("key1")
		if va[i] != va[0] {
			t.Fatalf("Value not match with input!")
		}
	}
	for i := 0; i < serverNum-deadNum; i++ {
		va[i] = agent[i].Get("key2")
		if va[i] != va[0] {
			t.Fatalf("Value not match with input!")
		}
	}
	fmt.Printf("  ... Passed\n")
	time.Sleep(1 * time.Second)
}

func TestDeadServer2(t *testing.T) {

	const serverNum = 7
	const deadNum = 3
	fmt.Printf("TestDeadServer Test 2:  test %d servers and %d of them is dead \n", serverNum, deadNum)

	var servers []server.Server = make([]server.Server, serverNum)
	var address []string = make([]string, serverNum)
	defer Close(servers)

	for i := 0; i < serverNum; i++ {
		address[i] = CreateAddress(i + 10)
	}
	for i := 0; i < serverNum-deadNum; i++ {
		servers[i], _ = server.NewServer(address, i, false, false)
	}

	var agent [serverNum]*FakeAgent
	for i := 0; i < serverNum-deadNum; i++ {
		agent[i] = MakeFakeAgent(servers)
	}

	for i := 0; i < serverNum-deadNum; i++ {
		agent[i].Put("key1", "value1")
		agent[i].Put("key2", "value2")
	}

	var va [serverNum]string
	for i := 0; i < serverNum-deadNum; i++ {
		va[i] = agent[i].Get("key1")
		if va[i] != va[0] {
			t.Fatalf("Value not match with input!")
		}
	}
	for i := 0; i < serverNum-deadNum; i++ {
		va[i] = agent[i].Get("key2")
		if va[i] != va[0] {
			t.Fatalf("Value not match with input!")
		}
	}
	fmt.Printf("  ... Passed\n")
	time.Sleep(1 * time.Second)
}

func TestPartition1(t *testing.T) {

	const groupServers = 2
	const groups = 2
	var servers_1 []server.Server = make([]server.Server, groupServers)
	var servers_2 []server.Server = make([]server.Server, groupServers)

	var address_1 []string = make([]string, groupServers*groups)
	var address_2 []string = make([]string, groupServers*groups)

	fmt.Printf("Partition Test1: 2 groups to test partition...\n")
	defer Close(servers_1)
	defer Close(servers_2)

	for i := 0; i < groupServers*2; i++ {
		address_1[i] = CreateAddress(i)
		address_2[i] = CreateAddress((i + groupServers) % (groupServers * groups))
	}

	for i := 0; i < groupServers; i++ {
		servers_1[i], _ = server.NewServer(address_1, i, false, false)
		servers_2[i], _ = server.NewServer(address_2, i, false, false)

	}

	agent_1 := MakeFakeAgent(servers_1)
	agent_2 := MakeFakeAgent(servers_2)

	agent_1.Put("group1", "value1")
	agent_2.Put("group2", "value2")

	agent_1.Assess(t, "group1", "value1")
	agent_2.Assess(t, "group1", "value1")

	agent_1.Assess(t, "group2", "value2")
	agent_2.Assess(t, "group2", "value2")

	fmt.Printf("  ... Passed\n")
	time.Sleep(1 * time.Second)
}

func TestPartition2(t *testing.T) {

	const groupServers = 3
	const groups = 3
	var servers_1 []server.Server = make([]server.Server, groupServers)
	var servers_2 []server.Server = make([]server.Server, groupServers)
	var servers_3 []server.Server = make([]server.Server, groupServers)

	var address_1 []string = make([]string, groupServers*groups)
	var address_2 []string = make([]string, groupServers*groups)
	var address_3 []string = make([]string, groupServers*groups)

	fmt.Printf("Partition Test2: 3 groups to test partition...\n")
	defer Close(servers_1)
	defer Close(servers_2)
	defer Close(servers_3)

	for i := 0; i < groupServers*2; i++ {
		address_1[i] = CreateAddress(i)
		address_2[i] = CreateAddress(i + groupServers)
		address_3[i] = CreateAddress((i + groupServers*2) % (groupServers * groups))
	}

	for i := 0; i < groupServers; i++ {
		servers_1[i], _ = server.NewServer(address_1, i, false, false)
		servers_2[i], _ = server.NewServer(address_2, i, false, false)
		servers_3[i], _ = server.NewServer(address_3, i, false, false)
	}

	agent_1 := MakeFakeAgent(servers_1)
	agent_2 := MakeFakeAgent(servers_2)
	agent_3 := MakeFakeAgent(servers_3)

	agent_1.Put("group1", "value1")
	agent_2.Put("group2", "value2")
	agent_3.Put("group3", "value3")

	agent_1.Assess(t, "group1", "value1")
	agent_2.Assess(t, "group1", "value1")
	agent_3.Assess(t, "group1", "value1")

	agent_1.Assess(t, "group2", "value2")
	agent_2.Assess(t, "group2", "value2")
	agent_3.Assess(t, "group2", "value2")

	agent_1.Assess(t, "group3", "value3")
	agent_2.Assess(t, "group3", "value3")
	agent_3.Assess(t, "group3", "value3")

	fmt.Printf("  ... Passed\n")
	time.Sleep(1 * time.Second)
}
