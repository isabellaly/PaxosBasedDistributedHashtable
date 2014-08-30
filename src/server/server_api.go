package server

type Server interface {
	Get(args *GetArgs, reply *GetReply) error
	Put(args *PutArgs, reply *PutReply) error
	Close()
	StorageSize() int
}
