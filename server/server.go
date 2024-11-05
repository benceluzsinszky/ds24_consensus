package main

import (
	proto "Chittychat/grpc"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

type Bully struct {
	proto.UnimplementedBullyServer
	nodeId int
	nodes  map[int]proto.Bully_ChatServer
	mu     sync.Mutex
}

func (s *Bully) broadcast(message string) {
	for _, client := range s.nodes {
		if err := client.Send(&proto.Message{Message: message}); err != nil {
			log.Printf("Failed to send message to node: %v", err)
		}
	}
}

func (s *Bully) Chat(stream proto.Bully_ChatServer) error {
	s.mu.Lock()
	newNodeId := s.nodeId
	s.nodeId++
	s.nodes[newNodeId] = stream
	s.mu.Unlock()

	joinMessage := fmt.Sprintf("%d joined", newNodeId)
	log.Println(joinMessage)
	s.broadcast(joinMessage)

	defer func() {
		s.mu.Lock()
		delete(s.nodes, newNodeId)
		s.mu.Unlock()

		leaveMessage := fmt.Sprintf("%d left", newNodeId)
		log.Println(leaveMessage)
		s.broadcast(leaveMessage)
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if !strings.Contains(in.Message, "timeout") {
			log.Println(fmt.Sprintf("Node %d: %s", newNodeId, in.Message))
		}
		s.broadcast(in.Message)
	}

}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Server listening on port 8080")

	s := grpc.NewServer()
	proto.RegisterBullyServer(s, &Bully{
		nodes: make(map[int]proto.Bully_ChatServer),
	})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
