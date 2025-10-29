// server.go
package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Message represents a chat message
type Message struct {
	Username string
	Text     string
	Time     time.Time
}

// Args for sending a message
type SendArgs struct {
	Msg Message
}

// Reply containing full history
type HistoryReply struct {
	History []Message
}

// ChatServer holds the messages and a mutex
type ChatServer struct {
	mu       sync.Mutex
	history  []Message
	maxStore int // optional: cap history if wanted
}

// SendMessage stores the message and returns full history
func (s *ChatServer) SendMessage(args SendArgs, reply *HistoryReply) error {
	if args.Msg.Text == "" {
		return errors.New("empty message")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// Append message
	s.history = append(s.history, args.Msg)

	// Return a copy of history
	out := make([]Message, len(s.history))
	copy(out, s.history)
	reply.History = out
	fmt.Printf("New message from %s: %s\n", args.Msg.Username, args.Msg.Text)
	return nil
}

// FetchHistory returns the current history without adding anything
func (s *ChatServer) FetchHistory(_ struct{}, reply *HistoryReply) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Message, len(s.history))
	copy(out, s.history)
	reply.History = out
	return nil
}

func main() {
	server := &ChatServer{
		history:  make([]Message, 0),
		maxStore: 1000,
	}

	rpc.Register(server)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		fmt.Println("listen error:", err)
		os.Exit(1)
	}
	fmt.Println("Chat RPC server listening on :1234")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		// Serve connection in a new goroutine
		go rpc.ServeConn(conn)
	}
}
