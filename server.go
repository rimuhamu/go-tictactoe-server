package main

import (
	"bufio"
	"fmt"
	"net"
)

type Client struct {
	conn net.Conn
	name string
	ch   chan string
}

var (
	clients  = make(map[net.Conn]*Client)
	joining  = make(chan *Client)
	leaving  = make(chan *Client)
	messages = make(chan string)
)

func broadcaster() {
	for {
		select {
		case msg := <-messages:
			for _, client := range clients {
				fmt.Fprintln(client.conn, msg)
			}
		case newClient := <-joining:
			clients[newClient.conn] = newClient
			fmt.Println("New client:", newClient.name)
		case leavingClient := <-leaving:
			fmt.Println("Leaving:", leavingClient.name)
			delete(clients, leavingClient.conn)
			close(leavingClient.ch)
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	client := &Client{
		conn: conn,
		name: conn.RemoteAddr().String(),
		ch:   make(chan string),
	}

	joining <- client

	messages <- fmt.Sprintf("New player joined: %s\n", client.name)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- fmt.Sprintf("%s: %s\n", client.name, input.Text())
	}

	leaving <- client
	messages <- fmt.Sprintf("Leaving: %s\n", client.name)
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on :8080")

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConn(conn)
	}

}
