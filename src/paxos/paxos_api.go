package paxos

type Paxos interface {
	StartPaxos(rid int, op interface{})
	GetLog(rid int) (bool, interface{})
	CommitFinished(opID int)
	Close()
}
