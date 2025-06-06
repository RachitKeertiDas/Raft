package main

import (
	"fmt"

	"github.com/RachitKeertiDas/raft/core"
	"github.com/RachitKeertiDas/raft/internal/config"
)

func main() {

	fmt.Println("Hello. Initializing Raft Server")

	//  Load configs
	serverConfigs, localConfigs := config.Init()
	fmt.Println(serverConfigs, localConfigs)

	// Initialize Server with Configs.

	// Initialize Logger

	// Initialize State Machine

	// Initalze Server
	// dummy code to keep the import, will change soon
	core.Init()
}
