package server

type RemoteStorageServer interface {
	Put(*PutArgs, *PutReply) error
	Get(*GetArgs, *GetReply) error
}

type ServerRPC struct {
	RemoteStorageServer
}

func Wrap(s RemoteStorageServer) RemoteStorageServer {
	return &ServerRPC{s}
}
