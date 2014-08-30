package server

type Request struct {
	AgentID   int
	RequestID int64
	Name      string
	Key       string
	Value     string
}

type GetArgs struct {
	AgentID   int
	RequestID int64
	Key       string
}

type GetReply struct {
	AgentID   int
	RequestID int64
	Value     string
	OK        bool
	Error     error 
}

type PutArgs struct {
	AgentID   int
	RequestID int64
	Key       string
	Value     string
}

type PutReply struct {
	AgentID   int
	RequestID int64
	OK        bool
	Error     error
}
