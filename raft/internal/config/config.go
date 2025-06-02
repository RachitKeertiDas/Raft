package config

import (
	"encoding/json"
	"fmt"
)

type GlobalConfig struct {
	numServers int
	RPCMethod  string
	serverInfo map[string]int
}

type LocalConfig struct {
	LogHandler          string
	StateMachineHandler string
}

func Init() (GlobalConfig, LocalConfig) {

	fmt.Println("Initializing Config Reader")
	a, _ := json.Marshal("1")
	fmt.Println(a)
	fmt.Println("reading configs from config file.")

	// for now just init globalConfigs locally
	// we will see about reading it from JSON later
	tempMap := make(map[string]int)
	clusterConfigs := GlobalConfig{5, "HTTPRPCHandler", tempMap}
	localConfigs := LocalConfig{"TextHandler", "NoActionHandler"}
	fmt.Println(clusterConfigs)

	return clusterConfigs, localConfigs
}
