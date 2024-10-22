package main

import (
	proto "Chittychat/grpc"
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// var opts []grpc.DialOption
	// conn, err := grpc.NewClient(*serverAddr,
	// if err != nil {
	// 	log.Fatalf("Failed to connect: %v", err)
	// }
	// defer conn.Close()

	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewChittyChatClient(conn)

	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	// Goroutine to send messages
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if err := stream.Send(&proto.Message{Message: scanner.Text()}); err != nil {
				log.Fatalf("Failed to send message: %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Fatalf("Failed to close send stream: %v", err)
		}
	}()

	// Receiving messages
	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Failed to receive message: %v", err)
		}
		fmt.Println("Received:", reply.Message)
	}
}

// stream, err := client.RecordRoute(context.Background())
// if err != nil {
//   log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
// }
// for _, point := range points {
//   if err := stream.Send(point); err != nil {
//     log.Fatalf("%v.Send(%v) = %v", stream, point, err)
//   }
// }
// reply, err := stream.CloseAndRecv()
// if err != nil {
//   log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
// }
// log.Printf("Route summary: %v", reply)
