package main

import (
	proto "Chittychat/grpc"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ChittyChat struct {
	proto.UnimplementedChittyChatServer
	messages []string
}

func (s *ChittyChat) Chat(stream proto.ChittyChat_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		fmt.Println("Received:", in.Message)
		s.messages = append(s.messages, in.Message)
		for _, msg := range s.messages {
			if err := stream.Send(&proto.Message{Message: msg}); err != nil {
				return err
			}
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	proto.RegisterChittyChatServer(s, &ChittyChat{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
