package paxos

import "net"
import "net/rpc"
//import "log"
import "sync"
import "time"
import "math/rand"

type paxos struct {
	phaseLock           sync.Mutex
	listen              net.Listener
	closed              bool
	nodes               []string
	self                int
	ops                 map[int]Operation
	maxNodeDone         map[int]int
	proposeLock         sync.Mutex
	callbackConnections map[string]*rpc.Client
	isDebug             bool
}

func max(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// major funcions

func NewPaxos(nodes []string, self int, rpcs *rpc.Server, isDebug bool) *paxos {
	px := &paxos{
		nodes:               nodes,
		self:                self,
		ops:                 make(map[int]Operation),
		maxNodeDone:         make(map[int]int),
		callbackConnections: make(map[string]*rpc.Client),
		isDebug:             isDebug,
	}
	for i, _ := range px.nodes {
		px.maxNodeDone[i] = -1
	}

	// if rpcs != nil {
	rpcs.RegisterName("Paxos", Wrap(px))
	// } else {
	// 	rpcs = rpc.NewServer()
	// 	rpcs.RegisterName("Paxos", Wrap(px))

	// 	l, err := net.Listen("tcp", ":"+nodes[self])
	// 	if err != nil {
	// 		log.Fatal("listen error: ", err)
	// 		return nil
	// 	}
	// 	px.listen = l
	// 	go func() {
	// 		for px.closed == false {
	// 			conn, err := px.listen.Accept()
	// 			if err == nil && px.closed == false {
	// 				go rpcs.ServeConn(conn)
	// 			} else if err == nil {
	// 				conn.Close()
	// 			}
	// 		}
	// 	}()
	// }
	return px
}

func (px *paxos) StartPaxos(opID int, v_a interface{}) {
	px.phaseLock.Lock()
	defer px.phaseLock.Unlock()

	if px.MinID() <= opID {
		opertaion := px.findOperation(opID)
		if opertaion.commited {
			return
		}
		go px.Propose(opID, v_a)
	} else {
	}
	return
}

func (px *paxos) GetLog(opID int) (bool, interface{}) {
	px.phaseLock.Lock()
	defer px.phaseLock.Unlock()

	if px.MinID() <= opID {
		operation := px.findOperation(opID)
		return operation.commited, operation.v_a
	}
	return false, nil
}

func (px *paxos) Propose(opID int, v_a interface{}) {
	px.proposeLock.Lock()
	defer px.proposeLock.Unlock()

	proposalNumber := -1
	nextProposal := -1
	self := px.self
	completed := false
	for !completed {
		nextProposal += 1
		proposalNumber = nextProposal
		indicator := len(px.nodes) / 2
		prepareAgreeCount := 0
		maxProposal := -1
		maxProposalValue := v_a

		paxosAgrs := &PaxosAgrs{opID, proposalNumber, px.maxNodeDone[self], self, nil}
		for _, node := range px.nodes {
			paxosReply := &PaxosReply{}
			if node == px.nodes[self] {
				px.Prepare(paxosAgrs, paxosReply)
			} else {
				hasNoErr := px.rpcCall(node, "Paxos.Prepare", paxosAgrs, paxosReply)
				if !hasNoErr {
					continue
				}
			}
			if paxosReply.OK {
				prepareAgreeCount++
				if paxosReply.N_a > maxProposal {
					maxProposal = paxosReply.N_a
					maxProposalValue = paxosReply.V_a
				}
			} else {
				nextProposal = max(nextProposal, paxosReply.N_h)
			}
		}
		acceptAgreeCount := 0
		paxosAgrs.V_a = maxProposalValue

		if prepareAgreeCount > indicator {

			for _, node := range px.nodes {
				paxosReply := &PaxosReply{}
				if node == px.nodes[self] {
					px.Accept(paxosAgrs, paxosReply)
				} else {
					hasNoErr := px.rpcCall(node, "Paxos.Accept", paxosAgrs, paxosReply)
					if !hasNoErr {
						continue
					}
				}
				if paxosReply.OK {
					acceptAgreeCount++
				} else {
					nextProposal = max(nextProposal, paxosReply.N_h)
				}
			}
		}
		if acceptAgreeCount > indicator {
			for _, node := range px.nodes {
				paxosReply := &PaxosReply{}
				if node == px.nodes[self] {
					px.Commit(paxosAgrs, paxosReply)
				} else {
					px.rpcCall(node, "Paxos.Commit", paxosAgrs, paxosReply)
				}
			}
			completed = true
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (px *paxos) Prepare(args *PaxosAgrs, reply *PaxosReply) error {
	px.phaseLock.Lock()
	defer px.phaseLock.Unlock()
	// synchronize operation between node
	px.maxNodeDone[args.Self] = max(px.maxNodeDone[args.Self], args.CommitFinished)
	px.clearLog()

	reply.OK = false
	operation := px.findOperation(args.Rid)
	if operation.n_h < args.Pid {
		operation.n_h = args.Pid

		px.ops[args.Rid] = operation
		reply.N_a = operation.n_a
		reply.V_a = operation.v_a
		reply.OK = true
	} else {
		reply.N_h = operation.n_h
	}
	return nil
}

func (px *paxos) Accept(args *PaxosAgrs, reply *PaxosReply) error {
	px.phaseLock.Lock()
	defer px.phaseLock.Unlock()

	reply.OK = false
	operation := px.findOperation(args.Rid)
	if operation.n_h <= args.Pid {
		operation.n_h = args.Pid
		operation.n_a = args.Pid

		operation.v_a = args.V_a
		reply.Pid = args.Pid
		px.ops[args.Rid] = operation
		reply.OK = true
	} else {
		reply.N_h = operation.n_h
	}
	return nil
}

func (px *paxos) Commit(args *PaxosAgrs, reply *PaxosReply) error {
	px.phaseLock.Lock()
	defer px.phaseLock.Unlock()
	operation := px.findOperation(args.Rid)
	operation.v_a = args.V_a
	operation.commited = true
	px.ops[args.Rid] = operation
	reply.OK = true
	return nil
}

func (px *paxos) CommitFinished(opID int) {
	px.phaseLock.Lock()
	defer px.phaseLock.Unlock()
	self := px.self
	px.maxNodeDone[self] = max(opID, px.maxNodeDone[self])
}

func (px *paxos) MaxID() int {
	maxOperationID := -1
	for opID := range px.ops {
		maxOperationID = max(maxOperationID, opID)
	}
	return maxOperationID
}

func (px *paxos) MinID() int {
	self := px.self
	minOperationID := px.maxNodeDone[self]
	for _, opDone := range px.maxNodeDone {
		minOperationID = min(minOperationID, opDone)
	}
	return minOperationID + 1
}

func (px *paxos) Close() {
	px.closed = true
	if px.listen != nil {
		px.listen.Close()
	}
}

// utility methods
func (px *paxos) rpcCall(address string, serviceMethod string, args interface{}, reply interface{}) bool {
	var err error
	if px.isDebug && rand.Int()%2 == 0 {
		return false
	}

	c, err := rpc.Dial("tcp", address)
	if err != nil {
		return false
	}
	defer c.Close()
	err = c.Call(serviceMethod, args, reply)
	if err == nil {
		return true
	} 
	return false
}

func (px *paxos) findOperation(opID int) Operation {
	if operation, found := px.ops[opID]; found {
		return operation
	} else {
		px.ops[opID] = CreateOperation()
		operation := px.ops[opID]
		return operation
	}
}

func (px *paxos) clearLog() {
	min := px.MinID()
	for opID := range px.ops {
		if opID < min {
			delete(px.ops, opID)
		}
	}
}
