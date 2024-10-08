package main

import (
	proto "Chittychat/grpc"
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ChittyChat struct {
	proto.UnimplementedChittyChatServer
	messages []string
}

func (s *ChittyChat) SendMessage(ctx context.Context, in *proto.Message) (*proto.Message, error) {
	fmt.Print(in)
	s.messages = append(s.messages, in.Message)
	return &proto.Message{Message: s.messages[0]}, nil
}

func (s *ChittyChat) Chat(ctx context.Context, in *proto.Message) (*proto.AllMessages, error) {
	fmt.Print(in)
	for {
		req, err := in
		s.messages = append(s.messages, in.Message)
	}
	return &proto.Message{Message: s.messages[0]}, nil
}

func main() {
	server := &ChittyChat{messages: []string{}}
	message := proto.Message{
		ParticipantName: "User123",
		Message:         "Hello, ChittyChat!",
		LamportTime:     1, // Example Lamport time (increment this based on your logic)
	}

	server.messages = append(server.messages, message.Message)

	server.start_server()
}

func (s *ChittyChat) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChittyChatServer(grpcServer, s)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}

}
