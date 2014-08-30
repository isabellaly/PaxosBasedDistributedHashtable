package paxos

type RemotePaxosServer interface {
	Prepare(*PaxosAgrs, *PaxosReply) error
	Accept(*PaxosAgrs, *PaxosReply) error
	Commit(*PaxosAgrs, *PaxosReply) error
}

type PaxosRPC struct {
	RemotePaxosServer
}

func Wrap(px RemotePaxosServer) RemotePaxosServer {
	return &PaxosRPC{px}
}
