package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"pbft-pra/utils"
	"time"
)

type Node struct {
	NodeId       string
	NodeTable    map[string]string
	CurrentState *State
	View         *View
	MsgBuffer    *MsgBuffer
	CommitedMsgs []*CommitedMsg
	// define three channels
	MsgEntrance chan interface{} // general chan can transfer any type data
	MsgDelivery chan interface{}
	Alarm       chan bool
}

type INode interface {
	NewNode(string) *Node
	dispatchMsg()
	routeMsg(interface{}) error
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
		make([]*CommitedMsg, 0),
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
			// if preprepared is done , do prepared
			preparedMsgs := node.appendPreMsgs(msg)
			node.clearMsgsBuffer(msg)
			node.MsgDelivery <- preparedMsgs
		}
	case *CommitedMsg:
		if node.CurrentState == nil || node.CurrentState.State != prepared {
			node.MsgBuffer.CommitedMsgs = append(node.MsgBuffer.CommitedMsgs, msg.(*CommitedMsg))
		} else {
			// if prepared is done ,commit
			CommittedMsgs := node.appendCommitedMsgs(msg)
			node.clearMsgsBuffer(msg)
			node.MsgDelivery <- CommittedMsgs
		}
	}
	return nil
}

// this is a protect thread , continuing send alarm to true
func (node *Node) alarmToDispatcher() {
	for {
		time.Sleep(utils.TimeDuration)
		node.Alarm <- true
	}
}

func (node *Node) resolveMsg() {
	errs := make([]error, 0)
	for {
		msgs := <-node.MsgDelivery
		switch msgs.(type) {
		case []*ReqMsg:
			errs = node.resolveRequestMsg(msgs.([]*ReqMsg))
			if errs != nil {
				for err := range errs {
					fmt.Println(err)
				}
			}
		case []*PrepreparedMsg:
			errs = node.resolvePreprepareMsg(msgs.([]*PrepreparedMsg))
			if errs != nil {
				for err := range errs {
					fmt.Println(err)
				}
			}
		}
	}
}

func (node *Node) resolvePreprepareMsg(msgs []*PrepreparedMsg) []error {
	errs := make([]error, 0)

	for _, msg := range msgs {
		err := node.GetPrepre(msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return errs
	}
	return nil
}

func (node *Node) GetPrepre(msg *PrepreparedMsg) error {
	LogMsg(msg) // print log
	err := node.createNewState()
	if err != nil {
		return err
	}
	// generate next
	preparedMsg, err := node.CurrentState.Preprepared(msg)
	if err != nil {
		return err
	}

	if preparedMsg != nil {
		preparedMsg.NodeId = node.NodeId
	}
}

func (node *Node) Preprepared(msg *PrepreparedMsg) error {

}

func (node *Node) resolveRequestMsg(msgs []*ReqMsg) []error {
	errs := make([]error, 0)
	for _, msg := range msgs {
		// generate the preprepared msg and broadcast
		err := node.GetReq(msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if errs != nil {
		return errs
	}
	return nil
}

// client send req ->
func (node *Node) GetReq(msg *ReqMsg) error {
	// consensus cope
	err := node.createNewState()
	if err != nil {
		return err
	}
	// begin to consensus
	PrepreparedMsg, err := node.CurrentState.StartConsensus(msg)
	if err != nil {
		return err
	}
	if PrepreparedMsg != nil {
		node.Boardcast(PrepreparedMsg, "/preprepared")
	}
	return nil
}

func (node *Node) createNewState() error {
	if node.CurrentState != nil {
		return errors.New("current consensus is not over")
	}
	var lastSeq int64
	if len(node.MsgBuffer.CommitedMsgs) == 0 {
		lastSeq = -1
	} else {
		lastSeq = node.MsgBuffer.CommitedMsgs[len(node.MsgBuffer.CommitedMsgs)-1].SequenceId
	}

	// create a new round consensus
	node.CurrentState = NewState(node.View.ViewId, lastSeq)

	return nil
}

func (node *Node) Boardcast(msg interface{}, path string) map[string]error {
	errorMap := make(map[string]error)

	for nodeId, url := range node.NodeTable {
		// avoid self
		if node.NodeId == node.NodeId {
			continue
		}

		jsonMsg, err := json.Marshal(msg) // to json
		if err != nil {
			errorMap[nodeId] = err
			continue
		}

		send(url+path, jsonMsg)

		if len(errorMap) == 0 {
			return nil
		} else {
			return errorMap
		}
	}
}

func send(url string, msg []byte) {
	buffer := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buffer)
}

func (node *Node) clearMsgsBuffer(msg interface{}) {
	// node.MsgBuffer.msg.(type) = make([]*msg.(type),0)
	switch msg.(type) {
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
