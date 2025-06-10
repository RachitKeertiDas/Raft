package core

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	state      string
	term       int
	votedFor   int
	logMachine LogMachineHandler
	rpcClient  RPCHandler
	selfData   ServerConfig
	nodes      []ServerConfig
}

type ServerConfig struct {
	Id      int
	Address string
}

type RPCHandler interface {
	RequestVote(int, int)
}

const electionTimeout = 150
const heartbeatTimeout = 5

func Init(logMachine LogMachineHandler, localConfig ServerConfig, nodes []ServerConfig) Server {
	// A server must initialize to a follower state.
	// I think in case logger gets it back, it should increment the term
	term, _ := logMachine.InitLogger()
	var rpcClient RPCHandler
	server := Server{"Follower", term, -1, logMachine, rpcClient, localConfig, nodes}
	return server
}

func (server *Server) Run() {
	// start the RPCListener
	// This will be done in a separate GoRoutine
	// This allows us to parallely check and transition to candidate if
	// te server times out.

	// go server.RPCListener.ListenforRequests()
	// TODO: Change to RPC Listener Method
	go server.listenforRequests()
	timeout := 2
	if server.state == "Follower" {
		for timeout < heartbeatTimeout && server.state == "Follower" {
			// Keep Listening to queries and waiting for timeout
			// Listening to queries will be done by the http handler.

		}
	}

	if server.state == "Candidate" {
		startElection(server)
	}

	if server.state == "Leader" {
		// logWrite Requests will be handled, replicated by the handler.
		// the handler will change the state accordingly
	}
}

func startElection(server *Server) {
	// To start an election, change state to candidate
	// increment term
	server.state = "Candidate"
	server.term += 1
	// channels := make([]chan int, 5)
	//  for idx, _ := range channels {
	// 	channels[idx] = make(chan int)
	// }

	for _, member := range server.nodes {
		if member.Id == server.selfData.Id {
			// don't send an RPC call to yourself
			// close the channel, we won't be getting anything
			continue
		}
		server.rpcClient.RequestVote(member.Id, server.term)
	}

}

func (server *Server) listenforRequests() {
	fmt.Println("Listening for HTTP Requests @ %s", server.selfData.Address)
	log.Fatal(http.ListenAndServe(server.selfData.Address, nil))
}
