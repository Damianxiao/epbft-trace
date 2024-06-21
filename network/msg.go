package network

type ReqMsg struct {
	TimeStamp  int64  `json:"timestamp"`
	ClientId   string `json:"clientID"`
	Operation  string `json:"operation"`
	SequenceId int64  `json:"sequenceID"`
}

type PrepreparedMsg struct {
	ViewId     int64   `json:"viewID"`
	SequenceId int64   `json:"sequenceID"`
	Digest     string  `json:"digest"`
	NodeId     string  `json:"NodeId"`
	RequestMsg *ReqMsg `json:"requestMsg"`
}

type PreparedMsg struct {
	ViewId     int64  `json:"viewID"`
	SequenceId int64  `json:"sequenceID"`
	Digest     string `json:"digest"`
	NodeId     string `json:"NodeId"`
}

type CommitedMsg struct {
	ViewId     int64  `json:"viewID"`
	SequenceId int64  `json:"sequenceID"`
	Digest     string `json:"digest"`
	NodeId     string `json:"NodeId"`
}

// a consensus round is complete over return to client
type ReplyMsg struct {
	ViewId    int64  `json:"viewID"`
	TimsStamp int64  `json:"timestamp"`
	ClientId  string `json:"clientID"`
	NodeId    string `json:"NodeId"`
	Result    string `json:"Result"`
}
