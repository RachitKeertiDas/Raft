package main

import (
    "fmt"
    "github.com/RachitKeertiDas/raft/internal/config"
)

func main(){
	fmt.Println("Hello. Initializing Raft Server")
	config.Init()

}
