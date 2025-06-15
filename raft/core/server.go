package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
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
	fmt.Print(server)
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

	// meanwhile timeout in the background in case we don't recieve anything
	timeout := 0
	for {
		if server.state == "Follower" {
			if timeout <= heartbeatTimeout {
				// Keep Listening to queries and waiting for timeout
				// Listening to queries will be done by the http handler.
				fmt.Printf("Time since Last Response :%d \n", timeout)
				time.Sleep(1 * time.Second)
				timeout += 1

			} else {
				fmt.Printf("No Heartbeat Recieved since %d. Time to convert to candidate.\n", timeout)
				server.state = "Candidate"
			}
		}

		if server.state == "Candidate" {
			// To start an election, increment term
			startElection(server)
		}

		if server.state == "Leader" {
			// logWrite Requests will be handled, replicated by the handler.
			// the handler will change the state accordingly
		}

	}
}

func startElection(server *Server) {
	// increment term
	server.term += 1

	// TODO: Randomize ephemeralTimeout
	ephemeralTimeout := electionTimeout
	fmt.Printf("Election timeout for Term %d by server %d is: %d\n", server.term, server.selfData.Id, ephemeralTimeout)

	// channels := make([]chan int, 5)
	//  for idx, _ := range channels {
	// 	channels[idx] = make(chan int)
	// }
	// for now let's consider sequential calss
	// will parallelize later
	numServers := len(server.nodes)

	responses := make([]int, numServers)
	// 0 is unknown
	for idx, _ := range responses {
		responses[idx] = -1
	}
	responses[server.selfData.Id-1] = 1 // vote for self

	for idx, member := range server.nodes {
		if member.Id == server.selfData.Id {
			// don't send an RPC call to yourself
			// close the channel, we won't be getting anything
			continue
		}
		var resp int = -1
		if server.state != "Candidate" {
			fmt.Println("Election is finished")
			break
		}

		if responses[idx] == -1 && server.state == "Candidate" {
			resp = server.requestVote(member.Address, server.term, ephemeralTimeout)
		}
		responses[idx] = resp

		// TODO: Refine and quit if responses not in favour
		sum := 0
		for _, vote := range responses {
			if vote >= 0 {
				sum += vote
			}
		}

		if sum*2 > numServers {
			fmt.Println("Majority Achieved")
			server.state = "Leader"
			return
		}
	}

}

func (server *Server) listenforRequests() {
	fmt.Printf("Listening for HTTP Requests @ %s \n", server.selfData.Address)
	http.HandleFunc("/voteRequest", requestVoteHandler)
	log.Fatal(http.ListenAndServe(server.selfData.Address, nil))
}

func (server *Server) requestVote(address string, term int, timeout int) int {
	// Create a httpClient with the specified timeout
	fmt.Printf("Address %s, term: %d, timeout %d", address, term, timeout)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := address + "/voteRequest"

	body := make(map[string]string)
	body["candidateTerm"] = strconv.Itoa(term)
	body["candidateId"] = strconv.Itoa(server.selfData.Id)

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Unexpected Error:%s \n", err)
	}
	fmt.Println(string(jsonBytes))
	bodyReader := bytes.NewReader(jsonBytes)

	resp, err := client.Post(url, "application/json", bodyReader)
	if err != nil {
		fmt.Printf("Unexpected Error:%s \n", err)
		return -1
	}
	respBody, err := io.ReadAll(resp.Body)
	fmt.Println(string(respBody))
	if err != nil {
		// the call didn't succeed. Return -1 so that it will be retried by the caller.
		return -1
	}
	fmt.Println("Decision arrives")
	return 1
}

func requestVoteHandler(w http.ResponseWriter, req *http.Request) {

	decision := 1
	fmt.Print(decision)
	responseBody := make(map[string]string)

	responseBody["decision"] = strconv.Itoa(decision)
	jsonBytes, _ := json.Marshal(responseBody)

	fmt.Fprintf(w, "%s", jsonBytes)
}
