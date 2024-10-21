package main

import (
	"fmt"

	"github.com/wilo087/go_chat/internal/server"
)

func main() {
	fmt.Println("Starting the chat server...")

	srv := server.NewServer(10, 50)
	srv.Start()
}
