package core

type Server struct {
	state string
	term  int
}

const electionTimeout = 150

func Init() Server {
	// A server must initialize to a follower state.
	// TODO: For now, initialize term to 0
	// I think in case logger gets it back, it should increment the term
	server := Server{"Follower", 0}
	return server
}

func Run(server *Server) {

	// keep waiting for queries until timeout

	var timeout = false
	if timeout == true {
		startElection(server)
	}

}

func startElection(server *Server) {
	// To start an election, change state to candidate
	// increment term
	server.state = "Candidate"
	server.term += 1
}
