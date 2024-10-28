package main

import (
	"bufio"
	"context"
	"log"
	"os"

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
	_, err = client.Join(context.Background(), &pb.JoinRequest{ParticipantId: participantID})
	if err != nil {
		log.Fatalf("Failed to join the chat: %v", err)
	}

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
			log.Printf("\"%s\" at Lamport Time: %d\n", msg.Message, msg.LamportTime)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if scanner.Scan() {
			message := scanner.Text()

			if message == "exit" {
				break
			}

			// Publish a message
			_, err := client.Publish(context.Background(), &pb.PublishRequest{
				ParticipantId: participantID,
				Message:       message,
			})
			if err != nil {
				log.Fatalf("Failed to publish message: %v", err)
			}
		}
	}

	// Leave the chat
	leaveResp, err := client.Leave(context.Background(), &pb.LeaveRequest{ParticipantId: participantID})
	if err != nil {
		log.Fatalf("Failed to leave the chat: %v", err)
	}
	log.Printf("%s\n", leaveResp.Message)

}
