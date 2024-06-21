package network

import (
	"errors"
	"fmt"
	"pbft-pra/utils"
)

type IState interface {
	NewState(int64, int64) *State
}

type State struct {
	ViewsId int64
	// logs
	MsgLogs        *MsgLogs
	State          Stage
	LastSequenceId int64
}

type MsgLogs struct {
	ReqMsg         *ReqMsg
	PreprepareMsgs map[string]*PrepreparedMsg
	PrepareMsgs    map[string]*PreparedMsg
	CommitMsgs     map[string]*CommitedMsg
}

type Stage int

// every state note that the state is over
const (
	Idle Stage = iota
	preprepared
	prepared
	committed
)

func NewState(viewId int64, lastSequenceId int64) *State {
	return &State{
		ViewsId:        viewId,
		LastSequenceId: lastSequenceId,
		State:          Idle,
	}
}

// 传入 req ,return preprepared start a new consensus
func (state *State) StartConsensus(req *ReqMsg) (*PrepreparedMsg, error) {
	state.MsgLogs.ReqMsg = req

	sequenceId := utils.NowTime()

	// lastSequenceId +1 as a new id
	if state.LastSequenceId != -1 {
		for state.LastSequenceId >= sequenceId {
			sequenceId += 1
		}
	}

	req.SequenceId = sequenceId

	// get req signature
	digest, err := utils.Digest(req)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// change the node state
	state.State = preprepared

	// construct the prepremsg
	return &PrepreparedMsg{
		ViewId:     state.ViewsId,
		SequenceId: sequenceId,
		Digest:     digest,
		RequestMsg: req,
	}, nil
}

func (state *State) Preprepared(msg *PrepreparedMsg) (*PreparedMsg, error) {
	state.MsgLogs.PreprepareMsgs[msg.NodeId] = msg
	if !state.verifyMsg(msg.ViewId, msg.SequenceId, msg.Digest) {
		return nil, errors.New("pre-prepared msg is error")
	}
	// change state and vote the choice

	state.State = preprepared

	// if node is honest
	return &PreparedMsg{
		ViewId:     state.ViewsId,
		SequenceId: msg.SequenceId,
		Digest:     msg.Digest,
	}, nil
}

func (state *State) Prepared(msg *PreparedMsg) (*CommitedMsg, error) {

	// verify the msg
	if !state.verifyMsg(msg.ViewId, msg.SequenceId, msg.Digest) {
		return nil, errors.New("msg is incorrect")
	}

	state.MsgLogs.PrepareMsgs[msg.NodeId] = msg

	// print the vote msg
	fmt.Printf("[Prepare-Vote]: %d\n", len(state.MsgLogs.PrepareMsgs))

	if state.IsPrepared() {
		state.State = committed

		return &CommitedMsg{
			ViewId:     state.ViewsId,
			SequenceId: msg.SequenceId,
			Digest:     msg.Digest,
		}, nil
	}

	return nil, nil

}

func (state *State) Committed(msg *CommitedMsg) (*ReplyMsg, *ReqMsg, error) {
	// end a round of consensus
	if !state.verifyMsg(msg.ViewId, msg.SequenceId, msg.Digest) {
		return nil, nil, errors.New("commit msg is incorrect")
	}

	state.MsgLogs.CommitMsgs[msg.NodeId] = msg

	fmt.Printf("[Commit-Vote]: %d\n", len(state.MsgLogs.CommitMsgs))

	if state.IsCommitted() {
		result := "new block has been submitted"

		state.State = committed

		return &ReplyMsg{
			ViewId:    state.ViewsId,
			TimsStamp: utils.NowTime(),
			ClientId:  state.MsgLogs.ReqMsg.ClientId,
			NodeId:    msg.NodeId,
			Result:    result,
		}, state.MsgLogs.ReqMsg, nil
	}

	return nil, nil, nil
}

func (state *State) IsPrepared() bool {
	if state.MsgLogs.PrepareMsgs == nil {
		return false
	}
	if len(state.MsgLogs.PrepareMsgs) < 2*utils.F {
		return false
	}
	// above 2/3
	return true
}

func (state *State) IsCommitted() bool {
	if state.MsgLogs.CommitMsgs == nil {
		return false
	}
	if len(state.MsgLogs.CommitMsgs) < 2*utils.F {
		return false
	}
	// above 2/3
	return true
}

func (state *State) verifyMsg(viewId int64, sequenceId int64, digest string) bool {
	// viewId should be now
	if state.ViewsId != viewId {
		return false
	}

	if state.LastSequenceId >= sequenceId {
		return false
	}
	return true

}
