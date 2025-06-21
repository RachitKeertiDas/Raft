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
	state         string
	term          int
	votedFor      int
	logMachine    LogMachineHandler
	rpcClient     RPCHandler
	selfData      ServerConfig
	nodes         []ServerConfig
	lastHeartbeat time.Time
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

const electionTimeout = 10
const heartbeatTimeout = time.Duration(10 * time.Second)

func Init(logMachine LogMachineHandler, localConfig ServerConfig, nodes []ServerConfig) Server {
	// A server must initialize to a follower state.
	// I think in case logger gets it back, it should increment the term
	term, _ := logMachine.InitLogger()
	var rpcClient RPCHandler
	server := Server{"Follower", term, -1, logMachine, rpcClient, localConfig, nodes, time.Now()}
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
	// timeout := 0
	for {
		currentTime := time.Now()
		if server.state == "Follower" {
			if currentTime.Sub(server.lastHeartbeat) <= heartbeatTimeout {
				// Keep Listening to queries and waiting for timeout
				// Listening to queries will be done by the http handler.
				fmt.Printf("Time since Last Response :%s \n", currentTime.Sub(server.lastHeartbeat))
				time.Sleep(1 * time.Second)
				//timeout += 1

			} else {
				fmt.Printf("No Heartbeat Recieved since %s. Time to convert to candidate.\n", currentTime.Sub(server.lastHeartbeat))
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
			server.lastHeartbeat = time.Now()
		}

		if server.state == "Leader" {
			// logWrite Requests will be handled, replicated by the handler.
			// the handler will change the state accordingly
			handleLeaderState(server)
		}

	}
}

func handleLeaderState(server *Server) {

	// InitializeLeaderState()

	// send HeartBeats for now.
	// get term from servers. if some one rejects
	// with a higher term - we should convert to a follower.

	for _, member := range server.nodes {
		if member.Id == server.selfData.Id {
			// don't send an RPC call to yourself
			// close the channel, we won't be getting anything
			server.lastHeartbeat = time.Now()
			continue
		}
		server.sendHeartbeat(member.Id, member.Address)
	}
	time.Sleep(1 * time.Second)
	// also need to send log entries - the below function will be changed
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
	voteRouteHandler := customHTTPHandler{server: server, customRouter: requestVoteHandler}
	heartbeatHandler := customHTTPHandler{server: server, customRouter: heartbeatHandler}
	http.Handle("/voteRequest", &voteRouteHandler)
	http.Handle("/heartbeat", &heartbeatHandler)
	log.Fatal(http.ListenAndServe(server.selfData.Address, nil))
}

func (server *Server) sendHeartbeat(id int, address string) {

	body := make(map[string]string)
	body["candidateTerm"] = strconv.Itoa(server.term)
	body["candidateId"] = strconv.Itoa(server.selfData.Id)
	url := address + "heartbeat"

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Unexpected Error:%s \n", err)
	}

	bodyReader := bytes.NewReader(jsonBytes)
	_, err = http.Post(url, "application/json", bodyReader)
	if err != nil {
		fmt.Println("Unexpected error occured", err, bodyReader)
	}
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

// TODO: rename to AppendEntries once logging functionality is integrated.
func heartbeatHandler(w http.ResponseWriter, req *http.Request, server *Server) {
	fmt.Println("Recieved Heartbeat request.")

	// in case everything is sucessful and server is follower.
	requestBody := make(map[string]string)
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Print("Unexpected error occured: %s", err)
	}
	json.Unmarshal(bytes, &requestBody)

	fmt.Println("[Heartbeat] Decoded RequestBody: ", requestBody)
	// recieving request body - see above

	if server.state == "Follower" {
		server.lastHeartbeat = time.Now()

	}
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
		// server has voted but not for the given candidate
		// hence reject the vote.
		if server.votedFor != -1 && server.votedFor != candidateId {
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
