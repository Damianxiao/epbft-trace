package network

import (
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
	ReqMsg      *ReqMsg
	PrepareMsgs map[string]*PreparedMsg
	CommitMsgs  map[string]*CommitedMsg
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

func (state *State) verifyMsg(viewId int64, sequenceId int64, digest string) bool {
	// viewId should be now
	if state.ViewsId != viewId {
		return false
	}

	if state.LastSequenceId >= sequenceId {
		return false
	}

}
