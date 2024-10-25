package main

import (
	"context"
	"log"
	"os"
	"sync"

	pb "chittychat/proto"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <participant_id>", os.Args[0])
	}

	participantID := os.Args[1]
	// Dial the server
	conn, err := grpc.DialContext(context.Background(), "localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChittyChatServiceClient(conn)
	

	// Join the chat
	joinResp, err := client.Join(context.Background(), &pb.JoinRequest{ParticipantId: participantID})
	if err != nil {
		log.Fatalf("Failed to join the chat: %v", err)
	}
	log.Printf("%s\n", joinResp.Message)

	// Start listening for broadcast messages in a separate goroutine
	go func() {
		stream, err := client.Broadcast(context.Background(), &pb.BroadcastMessage{ParticipantId: participantID})
		if err != nil {
			log.Printf("Failed to start broadcast stream: %v", err)
		}
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Printf("Error receiving message: %v", err)
			}
			log.Printf("%s\n", msg.Message)
		}
	}()
	
	// Publish a message
	publishResp, err := client.Publish(context.Background(), &pb.PublishRequest{
		ParticipantId: participantID,
		Message:       "Hello, Chitty-Chat!",
	})
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Printf("%s\n", publishResp.Message)

	// Leave the chat
	leaveResp, err := client.Leave(context.Background(), &pb.LeaveRequest{ParticipantId: participantID})
	if err != nil {
		log.Fatalf("Failed to leave the chat: %v", err)
	}
	log.Printf("%s\n", leaveResp.Message)

	// Wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-context.Background().Done()
	}()
	wg.Wait()
}
