package main

import (
	proto "Chittychat/grpc"
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf(err.Error())
	}

	client := proto.NewChittyChatClient(conn)

	message, err := client.SendMessage(context.Background(), &proto.Message{ParticipantName: "Lucian", Message: "Hello from the outside", LamportTime: 123})
	if err != nil {
		log.Fatalf(err.Error())
	}

	println(" - " + message.Message)
}
