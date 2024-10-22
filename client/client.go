package main

import (
	proto "Chittychat/grpc"
	"bufio"
	"context"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func isMessageValid(message string) bool {
	return len(message) < 128
}

func main() {

	lamport := 0

	conn, err := grpc.NewClient("0.0.0.0:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewChittyChatClient(conn)

	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if !isMessageValid(scanner.Text()) {
				log.Println("Message is too long, please keep it under 128 characters")
				continue
			}
			lamport++
			if err := stream.Send(&proto.Message{Message: scanner.Text(), LamportTime: int64(lamport)}); err != nil {
				log.Fatalf("Failed to send message: %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("Failed to close send stream: %v", err)
		}
	}()

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Failed to receive message: %v", err)
		}
		lamport = max(lamport, int(reply.LamportTime)) + 1

		log.Println(reply.Message)
	}
}
