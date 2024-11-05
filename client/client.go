package main

import (
	proto "Chittychat/grpc"
	"context"
	"io"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	var bullyId int
	var nodeId int

	// use stream because we are madlads

	conn, err := grpc.NewClient("0.0.0.0:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewBullyClient(conn)

	stream, err := client.Chat(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	// initial join, Node gets its ID from the server
	initialMessage, err := stream.Recv()
	if err != nil {
		log.Fatalf("Failed to receive message: %v", err)
	}
	if !strings.Contains(initialMessage.Message, "timeout") {
		log.Println(initialMessage.Message)
	}
	nodeId, err = strconv.Atoi(strings.Split(initialMessage.Message, " ")[0])
	bullyId = nodeId

	// when Node joins it initiantes bully algorithm
	bully(stream, nodeId, &bullyId)

	// goroutine too loop accessing critical section
	go useCriticalSection(stream, nodeId)

	// main thread
	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Failed to receive message: %v", err)
		}
		// if healthcheck shows somebody left, trigger bully algorithm
		if strings.Contains(reply.Message, "left") {
			bully(stream, nodeId, &bullyId)
		}
		// skip message logs if they are trash
		if !strings.Contains(reply.Message, "timeout") && !strings.Contains(reply.Message, "CS") && !strings.Contains(reply.Message, "Bully") {
			log.Println(reply.Message)
		}
		// if i am the bully and somebody is requesting to access critical section (CS), let them
		if strings.Contains(reply.Message, "CS") && nodeId == bullyId {
			senderNodeId, err := strconv.Atoi(strings.Split(reply.Message, " ")[1])
			if err != nil {
				log.Fatalf("Failed to convert string to int: %v", err)
			}
			criticalSection(bullyId, senderNodeId)
			// if a node is broadcasting coordinate trigger coordination
		} else if strings.Split(reply.Message, "(")[0] == "coordinate" {
			coordinate(stream, reply.Message, nodeId, &bullyId)
		}
	}

}

func bully(stream proto.Bully_ChatClient, nodeId int, bullyId *int) {
	// trigger coordination on all other nodes
	message := "coordinate(" + strconv.Itoa(nodeId) + ")"
	sendMessage(stream, message)

	// timeout for nodes to reply
	timeout := time.After(1 * time.Second)
	ticker := time.Tick(100 * time.Millisecond)

	for {
		select {
		// if nobody sends coordinate, it means this node has highest id and it is the bully
		case <-timeout:
			log.Printf("Node %d: I am the bully\n", nodeId)
			*bullyId = nodeId
			return
		case <-ticker:
			reply, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatalf("Failed to receive message: %v", err)
			}

			if strings.Split(reply.Message, "(")[0] == "coordinate" {
				log.Println("Received coordinate message")
				newBullyId, err := strconv.Atoi(strings.Split(strings.Split(reply.Message, "(")[1], ")")[0])
				if err != nil {
					log.Fatalf("Failed to convert string to int: %v", err)
				}
				// if coordinate message comes in with higher id, it means this node is not the bully, update global bully id
				if newBullyId > nodeId {
					log.Println("New bully is: ", newBullyId)
					*bullyId = newBullyId
					return
				}
			}
			// timeout message, because recv doesnt have inbuilt timeout because it is cancer
			message := "waiting for timeout"
			sendMessage(stream, message)
		}
	}
}

func coordinate(stream proto.Bully_ChatClient, message string, nodeId int, bullyId *int) {
	newBullyId, err := strconv.Atoi(strings.Split(strings.Split(message, "(")[1], ")")[0])
	if err != nil {
		log.Fatalf("Failed to convert string to int: %v", err)
	}
	// if coordinate message is by this node, return
	if newBullyId == nodeId {
		return
		// if new bully id is higher than current bully id, update global bully id
	} else if newBullyId > nodeId {
		log.Println("New bully is: ", newBullyId)
		*bullyId = newBullyId
	} else {
		// if this node has higher id then the one received initiate bullying
		bully(stream, nodeId, bullyId)
	}
}

func criticalSection(bullyId int, senderNodeId int) {
	// for the bully, if someone wants to access the cs
	message := "Bully " + strconv.Itoa(bullyId) + ": Critical Section updated by " + strconv.Itoa(senderNodeId)
	log.Println(message)

	return
}

func sendMessage(stream proto.Bully_ChatClient, message string) {
	// helper to send messages
	if err := stream.Send(&proto.Message{Message: message}); err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
}

func useCriticalSection(stream proto.Bully_ChatClient, nodeId int) {
	// each node should try to access the cs on random intervals between 3-10 seconds
	for {
		randomTime := time.Duration(rand.IntN(8)+3) * time.Second
		time.Sleep(randomTime)
		log.Println("Node " + strconv.Itoa(nodeId) + ": Requesting critical section")
		sendMessage(stream, ("CS " + strconv.Itoa(nodeId)))
		continue
	}
}
