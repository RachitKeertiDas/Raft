package main

import (
	"fmt"

	"github.com/RachitKeertiDas/raft/internal/config"
)

func main() {

	fmt.Println("Hello. Initializing Raft Server")

	//  Load configs
	serverConfigs, localConfigs := config.Init()
	fmt.Println(serverConfigs, localConfigs)
}
