package network

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Server struct {
	url  string
	node *Node
}

func NewServer(nodeId string) *Server {
	node := NewNode(nodeId)
	server := &Server{node.NodeTable[nodeId], node}
	server.setRoute()
	return server
}

func (server *Server) setRoute() {
	http.HandleFunc("/req", server.getReq)
	http.HandleFunc("/preprepare", server.getPrePrepare)
	http.HandleFunc("/prepare", server.getPrepare)
	http.HandleFunc("/commit", server.getCommit)
	http.HandleFunc("/reply", server.getReply)
}

func (server *Server) getReq(w http.ResponseWriter, r *http.Request) {
	var msg ReqMsg
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPrePrepare(w http.ResponseWriter, r *http.Request) {
	var msg PrepreparedMsg
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPrepare(w http.ResponseWriter, r *http.Request) {
	var msg PreparedMsg
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getCommit(w http.ResponseWriter, r *http.Request) {
	var msg CommitedMsg
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getReply(w http.ResponseWriter, r *http.Request) {
	var msg ReplyMsg
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	// server.node.GetReply(&msg)
}

func (server *Server) Start() {
	fmt.Printf("server will be start at %s", server.url)
	if err := http.ListenAndServe(server.url, nil); err != nil {
		fmt.Println(err)
		return
	}
}
