package network

import (
	"fmt"
)

func LogMsg(msg interface{}) {
	switch msg.(type) {
	case *ReqMsg:
		reqMsg := msg.(*ReqMsg)
		fmt.Printf("[REQUEST] ClientID: %s, Timestamp: %d, Operation: %s\n", reqMsg.ClientId, reqMsg.TimeStamp, reqMsg.Operation)
	case *PrepreparedMsg:
		prePrepareMsg := msg.(*PrepreparedMsg)
		fmt.Printf("[PREPREPARE] ViewId: %s, SequenceID: %d\n", prePrepareMsg.ViewId, prePrepareMsg.SequenceId)
	case *PreparedMsg:
		preparedMsg := msg.(*PreparedMsg)
		fmt.Printf("[PREPARE] NodeId: %s", preparedMsg.NodeId)
	case *CommitedMsg:
		CommitedMsg := msg.(*CommitedMsg)
		fmt.Printf("[Commit] NodeId: %s", CommitedMsg.NodeId)
	}
}
