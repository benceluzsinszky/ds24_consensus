package main

import (
	proto "Chittychat/grpc"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
)

type ChittyChat struct {
	proto.UnimplementedChittyChatServer
	messages []string
	clientId int
	lamport  int64
	clients  map[int]proto.ChittyChat_ChatServer
	mu       sync.Mutex
}

func (s *ChittyChat) broadcast(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lamport++
	for _, client := range s.clients {
		if err := client.Send(&proto.Message{Message: message, LamportTime: s.lamport}); err != nil {
			log.Printf("Failed to send message to client: %v", err)
		}
	}
}

func (s *ChittyChat) Chat(stream proto.ChittyChat_ChatServer) error {
	s.mu.Lock()
	newClientId := s.clientId
	s.clientId++
	s.clients[newClientId] = stream
	s.mu.Unlock()

	s.broadcast(fmt.Sprintf("Participant %d joined Chitty-Chat at Lamport time %d", newClientId, s.lamport))

	defer func() {
		s.mu.Lock()
		delete(s.clients, newClientId)
		s.mu.Unlock()
		s.broadcast(fmt.Sprintf("Participant %d left Chitty-Chat at Lamport time %d", newClientId, s.lamport))
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.lamport = max(s.lamport, in.LamportTime) + 1
		savedMessage := fmt.Sprintf("Client %d: %s at Lamport Time %d", newClientId, in.Message, s.lamport)
		s.messages = append(s.messages, savedMessage)
		s.mu.Unlock()

		// bigMessage := ""
		// for _, msg := range s.messages {
		// 	bigMessage += msg + "\n"
		// }

		// s.broadcast(bigMessage)
		s.broadcast(savedMessage)
	}

}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Server listening on port 8080")

	s := grpc.NewServer()
	proto.RegisterChittyChatServer(s, &ChittyChat{
		clients: make(map[int]proto.ChittyChat_ChatServer),
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
