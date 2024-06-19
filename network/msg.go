package network

type ReqMsg struct {
	TimeStamp  int64
	ClientId   string
	Operation  string
	SequenceId int64
}

type PrepreparedMsg struct {
	ViewSId    int64
	SequenceId int64
	Digest     string // means signature
	RequestMsg *ReqMsg
}

type PreparedMsg struct {
	ViewId     int64
	SequenceId int64
	Digest     string
	NodeId     string
}

type CommitedMsg struct {
	ViewId     int64
	SequenceId int64
	Digest     string
	NodeId     string
}

// a consensus round is complete over return to client
type ReplyMsg struct {
	ViewId    int64
	TimsStamp int64
	ClientId  string
	NodeId    string
	Result    string
}
