package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"

	pb "chittychat/proto"

	"google.golang.org/grpc"
)

type ChittyChatService struct {
	pb.UnimplementedChittyChatServiceServer

	mu           sync.Mutex
	lamport_time int64
	participants map[string]chan *pb.BroadcastMessage
}

func main() {
	// Set up a listener on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register ChittyChatService with the gRPC server
	pb.RegisterChittyChatServiceServer(grpcServer, &ChittyChatService{
		participants: make(map[string]chan *pb.BroadcastMessage),
	})

	// Simulate opening multiple clients
	for i := 0; i < 3; i++ { // Open 3 clients
		username := "User" + strconv.Itoa(i+1)
		err := openClient(username)
		if err != nil {
			fmt.Printf("Error opening client: %v\n", err)
		}
	}

	// Log server start and start serving
	log.Printf("Server is listening on %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// The Join() method handles the logic for participants joining the chat.

func (s *ChittyChatService) Join(c context.Context, req *pb.JoinRequest) (*pb.JoinResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lamport_time++

	if _, exists := s.participants[req.ParticipantId]; exists {
		return nil, fmt.Errorf("participant %s already exists", req.ParticipantId)
	}

	s.participants[req.ParticipantId] = make(chan *pb.BroadcastMessage, 10)

	joinMessage := &pb.BroadcastMessage{
		ParticipantId: req.ParticipantId,
		Message:       fmt.Sprintf("Participant %s joined Chitty-Chat", req.ParticipantId),
		LamportTime:   s.lamport_time,
	}

	s.broadcast(joinMessage) //Broadcast the joinMessage to all participants using the broadcast() method

	return &pb.JoinResponse{ //Returns the joinResponse to the joining client, to know that the join was succesful
		Message:     joinMessage.Message,
		LamportTime: s.lamport_time,
	}, nil
}

// The Publish() method handles the logic for participants publishing messages to chitty-chat.
func (s *ChittyChatService) Publish(c context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lamport_time++

	if _, exists := s.participants[req.ParticipantId]; !exists {
		return nil, fmt.Errorf("participant %s does not exists", req.ParticipantId)
	}

	publishMessage := &pb.BroadcastMessage{
		ParticipantId: req.ParticipantId,
		Message:       req.Message,
		LamportTime:   s.lamport_time,
	}

	s.broadcast(publishMessage) //Publish the publishMessage to all participants

	return &pb.PublishResponse{ //Returns the publishResponse to the client publishing the message, to know that the publish was succesful
		Message:     "Message succesfully published",
		LamportTime: s.lamport_time,
	}, nil

}

// The Leave() method handles the logic for participants leaving chitty-chat.

func (s *ChittyChatService) Leave(c context.Context, req *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lamport_time++

	if _, exists := s.participants[req.ParticipantId]; !exists {
		fmt.Println("In the Leave methods if statement")
		return nil, fmt.Errorf("participant %s does not exists", req.ParticipantId)
	}

	delete(s.participants, req.ParticipantId) //Deletes the leaving participant and the corresponding channel value from the map

	leaveMessage := &pb.BroadcastMessage{
		ParticipantId: req.ParticipantId,
		Message:       fmt.Sprintf("Participant %s has left Chitty-Chat at Lamport time %d", req.ParticipantId, s.lamport_time),
		LamportTime:   s.lamport_time,
	}

	s.broadcast(leaveMessage) //Broadcast the leaveMessage to all participants using the broadcast() method

	return &pb.LeaveResponse{ //Returns the leaveMessage to the leaving client, to know that the leave was succesful
		Message:     leaveMessage.Message,
		LamportTime: s.lamport_time,
	}, nil
}

func (s *ChittyChatService) Broadcast(msg *pb.BroadcastMessage, grpc grpc.ServerStreamingServer[pb.BroadcastMessage]) error {
	participantId := msg.ParticipantId
	s.mu.Lock()
	msgChannel, exists := s.participants[participantId]
	s.mu.Unlock()

	if !exists {
		return fmt.Errorf("participant %s does not exists", participantId)
	}

	for {
		select {
		case msg := <-msgChannel:
			if err := grpc.Send(msg); err != nil {
				log.Printf("Failed to send message to %s: %v", participantId, err)
				return err
			}
		case <-grpc.Context().Done():
			log.Printf("Stream for participant %s closed", participantId)
			return nil
		}
	}
}

func (s *ChittyChatService) broadcast(msg *pb.BroadcastMessage) {
	for _, ch := range s.participants {
		ch <- msg
	}
}

func openClient(username string) error {
	var cmd *exec.Cmd

	// Get the current directory (the one from which the server was run)
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current directory: %v", err)
	}

	// Define the command to change to the Client directory and run the client with the username
	clientDir := workingDir + "/../Client" // Adjust this path as needed based on your project structure

	// Detect the operating system
	switch runtime.GOOS {
	case "linux":
		// Open a new terminal and run the client on Linux (gnome-terminal)
		cmd = exec.Command("gnome-terminal", "--", "go", "run", fmt.Sprintf("cd %s && go run chittychatclient.go %s", clientDir, username))
	case "darwin":
		// Open a new terminal and run the client on macOS (osascript)
		cmd = exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to do script "cd %s && go run chittychatclient.go %s"`, clientDir, username))
	case "windows":
		// Open a new terminal and run the client on Windows (start command)
		cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", fmt.Sprintf("cd %s && go run chittychatclient.go %s", clientDir, username))
	default:
		return fmt.Errorf("unsupported platform")
	}

	// Start the command
	return cmd.Start()
}
