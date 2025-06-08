package config

import (
	"encoding/json"
	"fmt"

	"github.com/RachitKeertiDas/raft/core"
)

type GlobalConfig struct {
	numServers int
	RPCMethod  string
	ServerInfo []core.ServerConfig
}

type LocalConfig struct {
	LogHandler          string
	StateMachineHandler string
	ServerInfo          core.ServerConfig
}

func Init() (GlobalConfig, LocalConfig) {

	fmt.Println("Initializing Config Reader")
	a, _ := json.Marshal("1")
	fmt.Println(a)
	fmt.Println("reading configs from config file.")

	// for now just init globalConfigs locally
	// we will see about reading it from JSON later
	localConfig := core.ServerConfig{3, "http://localhost:9073/"}
	clusterNodes := [5]core.ServerConfig{{1, "http://localhost:9071/"}, {2, "http://localhost:9072/"}, {3, "http://localhost:9073/"}, {4, "http://localhost:9074/"}, {5, "http://localhost:9075/"}}

	clusterConfigs := GlobalConfig{5, "HTTPRPCHandler", clusterNodes[:]}

	localConfigs := LocalConfig{"csvLogger", "NoActionHandler", localConfig}
	fmt.Println(clusterConfigs)

	return clusterConfigs, localConfigs
}
