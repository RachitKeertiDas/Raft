package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

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

	fmt.Println("Initializing Config Reader.")
	fileName := "localConfig.json"
	fmt.Printf("Reading configs from config file %s...", fileName)

	// TODO: Replace with common util function
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error opening File: %s", err)
		return GlobalConfig{}, LocalConfig{}
	}
	defer f.Close()

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		fmt.Println("Error Converting File to bytes: %s", err)
		return GlobalConfig{}, LocalConfig{}
	}
	fmt.Printf("File Data: %s\n", string(fileBytes))

	var localConfig LocalConfig
	json.Unmarshal(fileBytes, &localConfig)
	fmt.Println("Printing local Configs:", localConfig)

	// for now just init globalConfigs locally
	// we will see about reading it from JSON later
	clusterNodes := [5]core.ServerConfig{{1, "http://localhost:9071/"}, {2, "http://localhost:9072/"}, {3, "http://localhost:9073/"}, {4, "http://localhost:9074/"}, {5, "http://localhost:9075/"}}

	clusterConfigs := GlobalConfig{5, "HTTPRPCHandler", clusterNodes[:]}

	fmt.Println(localConfig)
	fmt.Println(clusterConfigs)

	return clusterConfigs, localConfig
}
