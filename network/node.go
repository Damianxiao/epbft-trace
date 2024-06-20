package network

type Node struct {
	NodeId       string
	NodeTable    map[string]string
	CurrentState *State
	View         *View
	MsgBuffer    *MsgBuffer
	// define three channels
	MsgEntrance chan interface{} // general chan can transfer any type data
	MsgDelivery chan interface{}
	Alarm       chan bool
}

type INode interface {
	NewNode(string) *Node
	dispatchMsg()
	alarmToDispatcher()
	resolveMsg()
}

type View struct {
	ViewId int64
	Leader string
}

type MsgBuffer struct {
	ReqMsgs         []*ReqMsg
	PrepreparedMsgs []*PrepreparedMsg
	PreparedMsgs    []*PreparedMsg
	CommitedMsgs    []*CommitedMsg
	ReplyMsgs       []*ReplyMsg
}

func NewNode(Id string) *Node {
	node := &Node{
		Id,
		map[string]string{
			"node1": "localhost:1111",
			"node2": "localhost:2222",
			"node3": "localhost:3333",
			"node4": "localhost:4444",
		},
		nil,
		&View{
			0,
			"node1",
		},
		&MsgBuffer{
			make([]*ReqMsg, 0),
			make([]*PrepreparedMsg, 0),
			make([]*PreparedMsg, 0),
			make([]*CommitedMsg, 0),
			make([]*ReplyMsg, 0),
		},
		// channels
		make(chan interface{}),
		make(chan interface{}),
		make(chan bool),
	}
	// start listening thread

	return node
}

func (node *Node) dispatchMsg() {
	for {
		select {
		// receive msg from chan
		case msg := <-node.MsgEntrance:
			err := node.routeMsg(msg)
			if err != nil {
				fmt.print(err)
			}
		case <-node.Alarm:
			err := node.routeMsgWhenAlarmed()
			if err != nil {
				fmt.print(err)
			}
		}

	}
}

func (node *Node) routeMsg(msg interface{}) error {
	// make different branch depends on type of msg
	switch msg.(type) { // interface{} . type
	case *ReqMsg:
		// if state is nil , then put into msg chan instantly, until state is not nil
		if node.CurrentState == nil {
			// new append a reqmsgs
			reqMsgs := node.appendReqMsgs(msg)
			// clear node reqMsgs pool
			node.clearMsgsBuffer(msg)
			// into channel
			node.MsgDelivery <- reqMsgs
		} else {
			node.MsgBuffer.ReqMsgs = append(node.MsgBuffer.ReqMsgs, msg.(*ReqMsg))
		}
	case *PrepreparedMsg:
		if node.CurrentState == nil {
			prepreparedMsgs := node.appendPrepreMsgs(msg)
			node.clearMsgsBuffer(msg)
			node.MsgDelivery <- prepreparedMsgs
		} else {
			node.MsgBuffer.PrepreparedMsgs = append(node.MsgBuffer.PrepreparedMsgs, msg.(*PrepreparedMsg))
		}
	case *PreparedMsg:
		if node.CurrentState == nil || node.CurrentState.State != preprepared {
			// if state is still nil or not in preprepared
			node.MsgBuffer.PreparedMsgs = append(node.MsgBuffer.PreparedMsgs, msg.(*PreparedMsg))
		} else {
			preparedMsgs := node.appendPreMsgs(msg)
			node.clearMsgsBuffer(msg)
			node.MsgDelivery <- preparedMsgs
		}
	case *CommitedMsg:
		if node.CurrentState == nil || node.CurrentState.State != prepared {
			node.MsgBuffer.CommitedMsgs = append(node.MsgBuffer.CommitedMsgs, msg.(*CommitedMsg))
		} else {
			// commit
			CommittedMsgs := node.appendCommitedMsgs(msg)
			node.clearMsgsBuffer(msg)
			node.MsgDelivery <- CommittedMsgs
		}
	}
	return nil
}

func (node *Node) clearMsgsBuffer(msgs interface{}) {
	// node.MsgBuffer.msg.(type) = make([]*msg.(type),0)
	switch msgs.(type) {
	case *ReqMsg:
		node.MsgBuffer.ReqMsgs = make([]*ReqMsg, 0)
	case *PrepreparedMsg:
		node.MsgBuffer.PrepreparedMsgs = make([]*PrepreparedMsg, 0)
	case *PreparedMsg:
		node.MsgBuffer.PreparedMsgs = make([]*PreparedMsg, 0)
	case *CommitedMsg:
		node.MsgBuffer.CommitedMsgs = make([]*CommitedMsg, 0)
	case *ReplyMsg:
		node.MsgBuffer.ReplyMsgs = make([]*ReplyMsg, 0)
	}
}

func (node *Node) appendReqMsgs(msg interface{}) []*ReqMsg {
	reqMsgs := make([]*ReqMsg, len(node.MsgBuffer.ReqMsgs))
	copy(reqMsgs, node.MsgBuffer.ReqMsgs)
	reqMsgs = append(reqMsgs, msg.(*ReqMsg))
	return reqMsgs
}

func (node *Node) appendPrepreMsgs(msg interface{}) []*PrepreparedMsg {
	PrepreparedMsgs := make([]*PrepreparedMsg, len(node.MsgBuffer.ReqMsgs))
	copy(PrepreparedMsgs, node.MsgBuffer.PrepreparedMsgs)
	PrepreparedMsgs = append(PrepreparedMsgs, msg.(*PrepreparedMsg))
	return PrepreparedMsgs
}

func (node *Node) appendPreMsgs(msg interface{}) []*PreparedMsg {
	PreparedMsgs := make([]*PreparedMsg, len(node.MsgBuffer.ReqMsgs))
	copy(PreparedMsgs, node.MsgBuffer.PreparedMsgs)
	PreparedMsgs = append(PreparedMsgs, msg.(*PreparedMsg))
	return PreparedMsgs
}

func (node *Node) appendCommitedMsgs(msg interface{}) []*CommitedMsg {
	CommitedMsgs := make([]*CommitedMsg, len(node.MsgBuffer.ReqMsgs))
	copy(CommitedMsgs, node.MsgBuffer.CommitedMsgs)
	CommitedMsgs = append(CommitedMsgs, msg.(*CommitedMsg))
	return CommitedMsgs
}
