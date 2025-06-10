package main

import (
	"fmt"

	"github.com/RachitKeertiDas/raft/core"
	"github.com/RachitKeertiDas/raft/internal/config"
	"github.com/RachitKeertiDas/raft/logMachine"
)

func main() {

	fmt.Println("Hello. Initializing Raft Server")

	//  Load configs
	serverConfigs, localConfigs := config.Init()
	fmt.Println(serverConfigs, localConfigs)

	// Initialize Server with Configs.

	// Initialize Logger
	logMachine := new(logMachine.CsvLogger)
	logMachine.InitConfig("FileLogs.csv")

	// Initialize State Machine

	// Initalize Server
	// dummy code to keep the import, will change soon

	server := core.Init(logMachine, localConfigs.ServerInfo, serverConfigs.ServerInfo)
	server.Run()
}
