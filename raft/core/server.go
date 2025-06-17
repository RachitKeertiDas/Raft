package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type customHTTPHandler struct {
	server       *Server
	customRouter func(w http.ResponseWriter, req *http.Request, server *Server)
}

func (httpHandler *customHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpHandler.customRouter(w, r, httpHandler.server)
}

const electionTimeout = 5
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
			// If election would have succeeded, state will be leader
			// if election would have timed out due to a split vote, state wold still be candidate
			// if election would have failed (someone else became leader), state will be follower
			// just set timeout to 0, as timeout only matters if state is follower.
			timeout = 0
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
	server.votedFor = server.selfData.Id

	// TODO: Randomize ephemeralTimeout
	ephemeralTimeout := electionTimeout
	ephemeralDuration := time.Duration(ephemeralTimeout) * time.Second
	fmt.Printf("Election timeout for Term %d by server %d is: %d\n", server.term, server.selfData.Id, ephemeralTimeout)

	// channels := make([]chan int, 5)
	//  for idx, _ := range channels {
	// 	channels[idx] = make(chan int)
	// }
	// for now let's consider sequential calss
	// will parallelize later
	numServers := len(server.nodes)
	timeStart := time.Now()
	fmt.Println("ELection started at timeStart", timeStart)

	responses := make([]int, numServers)
	// 0 is unknown
	for idx, _ := range responses {
		responses[idx] = -1
	}
	responses[server.selfData.Id-1] = 1 // vote for self

	for time.Now().Sub(timeStart) < ephemeralDuration {

		time.Sleep(1 * time.Second)
		fmt.Println("New Loooop\n\n")
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
				//	fmt.Println("Recieved Response", resp)
			}
			responses[idx] = resp

			// TODO: Refine and quit if responses not in favour
			accepts := 0
			for _, vote := range responses {
				if vote == 1 {
					accepts += 1
				}
			}

			rejects := 0
			for _, vote := range responses {
				if vote == 0 {
					rejects += 1
				}
			}

			if accepts*2 > numServers {
				fmt.Println("Majority Achieved")
				server.state = "Leader"
				return
			} else if rejects*2 > numServers {
				fmt.Println("Rejected election quest. Convert back to follower.")
				server.state = "Follower"
				return
			}
		}

	}
}

func (server *Server) listenforRequests() {
	fmt.Printf("Listening for HTTP Requests @ %s \n", server.selfData.Address)
	routeHandler := customHTTPHandler{server: server, customRouter: requestVoteHandler}
	http.Handle("/voteRequest", &routeHandler)
	log.Fatal(http.ListenAndServe(server.selfData.Address, nil))
}

func (server *Server) requestVote(address string, term int, timeout int) int {
	// Create a httpClient with the specified timeout
	fmt.Printf("Address %s, term: %d, timeout %d", address, term, timeout)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := address + "voteRequest"

	body := make(map[string]string)
	body["candidateTerm"] = strconv.Itoa(term)
	body["candidateId"] = strconv.Itoa(server.selfData.Id)

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Unexpected Error:%s \n", err)
	}
	fmt.Println(string(jsonBytes))
	bodyReader := bytes.NewReader(jsonBytes)
	fmt.Println("Making a call to", url)
	resp, err := client.Post(url, "application/json", bodyReader)
	if err != nil {
		fmt.Printf("Unexpected Error:%s \n", err)
		return -1
	}
	respBody := make(map[string]string)
	respBytes, err := io.ReadAll(resp.Body)
	fmt.Println("Decision from server ", string(respBytes))
	if err != nil {
		// the call didn't succeed. Return -1 so that it will be retried by the caller.
		return -1
	}
	json.Unmarshal(respBytes, &respBody)
	decision, err := strconv.Atoi(respBody["decision"])
	if err != nil {
		fmt.Println("Unexpected error ", err)
	}
	return decision
}

func requestVoteHandler(w http.ResponseWriter, req *http.Request, server *Server) {

	// fmt.Println("Server Term @ host:", server.term, req.Header, req.Method)
	fmt.Println("[VoteHandler]:Server Term @ host:", server.term)

	requestBody := make(map[string]string)
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Print("Unexpected error occured: %s", err)
	}
	json.Unmarshal(bytes, &requestBody)
	decision := 1

	fmt.Println("[VoteHandler]Decoded RequestBody: ", requestBody)
	if val, ok := requestBody["candidateTerm"]; ok {
		fmt.Println("Candidate Term", val)
	}
	candidateTerm := requestBody["candidateTerm"]
	candidateId, _ := strconv.Atoi(requestBody["candidateId"])
	candidateTermInt, _ := strconv.Atoi(candidateTerm)
	if candidateTermInt < server.term {
		// reject since the candidate is behind
		fmt.Printf("[VoteHandler]Rejecting Decision since %d < %d", candidateTerm, server.term)
		decision = 0
	} else if candidateTermInt == server.term {
		if server.votedFor != 1 && server.votedFor != candidateId {
			decision = 0
		}
	}

	fmt.Println(decision)
	responseBody := make(map[string]string)

	responseBody["decision"] = strconv.Itoa(decision)
	responseBody["serverTerm"] = strconv.Itoa(server.term)

	jsonBytes, _ := json.Marshal(responseBody)

	fmt.Fprintf(w, "%s", jsonBytes)
}
