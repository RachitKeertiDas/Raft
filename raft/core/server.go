package core

type Server struct {
	state    string
	term     int
	votedFor int
}

const electionTimeout = 150
const heartbeatTimeout = 5

func Init() Server {
	// A server must initialize to a follower state.
	// TODO: For now, initialize term to 0
	// I think in case logger gets it back, it should increment the term
	server := Server{"Follower", 0, -1}
	return server
}

func Run(server *Server) {
	// start the RPCListener
	// This will be done in a separate GoRoutine
	// This allows us to parallely check and transition to candidate if
	// te server times out.

	// go RPCListener.ListenForWriteRequests()
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

	// server.RPCHandler.RequestVote()

	// RPCServer.
}
