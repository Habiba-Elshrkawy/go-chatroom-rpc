// client.go
package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type Message struct {
	Username string
	Text     string
	Time     time.Time
}

type SendArgs struct {
	Msg Message
}

type HistoryReply struct {
	History []Message
}

func printHistory(h []Message) {
	if len(h) == 0 {
		fmt.Println("[No messages yet]")
		return
	}
	fmt.Println("----- Chat history -----")
	for i, m := range h {
		fmt.Printf("%d) [%s] %s: %s\n", i+1, m.Time.Format("2006-01-02 15:04:05"), m.Username, m.Text)
	}
	fmt.Println("------------------------")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run client.go <username>")
		return
	}
	username := os.Args[1]

	// connect to RPC server
	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer client.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Hello %s! Type messages and press Enter. Type 'exit' to quit. Type '/history' to fetch history.\n", username)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("read error:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "exit" {
			fmt.Println("bye!")
			return
		}
		if line == "" {
			continue
		}

		// If user asks only for history
		if line == "/history" {
			var histReply HistoryReply
			err = client.Call("ChatServer.FetchHistory", struct{}{}, &histReply)
			if err != nil {
				fmt.Println("RPC error (FetchHistory):", err)
				continue
			}
			printHistory(histReply.History)
			continue
		}

		// Otherwise send the message
		args := SendArgs{
			Msg: Message{
				Username: username,
				Text:     line,
				Time:     time.Now(),
			},
		}
		var reply HistoryReply
		err = client.Call("ChatServer.SendMessage", args, &reply)
		if err != nil {
			fmt.Println("RPC error (SendMessage):", err)
			// Optionally attempt to reconnect once
			// trying to reconnect
			fmt.Println("attempting to reconnect...")
			client, err = rpc.Dial("tcp", "127.0.0.1:1234")
			if err != nil {
				fmt.Println("reconnect failed:", err)
				return
			}
			// retry sending once after reconnect
			err = client.Call("ChatServer.SendMessage", args, &reply)
			if err != nil {
				fmt.Println("send failed after reconnect:", err)
				return
			}
		}
		// Print history returned by SendMessage
		printHistory(reply.History)
	}
}
