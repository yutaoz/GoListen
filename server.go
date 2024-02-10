package main

import (
	"fmt"
	"net"
)

func handleConn(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr()

	fmt.Println("Client connected from:", remoteAddr)
}

func main() {
	listener, err := net.Listen("tcp", ":8993")
	if err != nil {
		fmt.Println("Error on server start")
		return
	}

	defer listener.Close()
	fmt.Println("Server listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConn(conn)
	}
}
