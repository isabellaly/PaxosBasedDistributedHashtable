package paxos

type Operation struct {
	n_a      int //highest Accepted
	n_h      int //hightest Proposal Number
	commited bool
	v_a      interface{} // each operation value
}

func CreateOperation() Operation {
	operation := &Operation{
		n_a:      -1, //highest Accepted
		n_h:      -1, //hightest Proposal Number
		commited: false,
		v_a:      nil, // each operation value
	}
	return *operation
}

type PaxosAgrs struct {
	Rid            int // operation id
	Pid            int // proposal number
	CommitFinished int
	Self           int // record self
	V_a            interface{}
}

type PaxosReply struct {
	N_a int //highest Accepted
	N_h int //hightest Proposal Number
	OK  bool
	Pid int         // Proposal number
	V_a interface{} // each operation value
}
