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

func (node *Node) routeMsg(msg interface{}) []error {
	// make different branch depends on type of msg
	switch msg {
	case ReqMsg:

	}
}
