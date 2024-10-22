package system

import (
	"context"
	"fmt"
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
		Message:       fmt.Sprintf("Participant %s joined Chitty-Chat at Lamport time %d", req.ParticipantId, s.lamport_time),
		LamportTime:   s.lamport_time,
	}

	s.broadcast(joinMessage, req.ParticipantId) //Broadcast the joinMessage to all participants using the broadcast() method

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

	s.broadcast(publishMessage, req.ParticipantId) //Publish the publishMessage to all participants

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
		return nil, fmt.Errorf("participant %s does not exists", req.ParticipantId)
	}

	delete(s.participants, req.ParticipantId) //Deletes the leaving participant and the corresponding channel value from the map

	leaveMessage := &pb.BroadcastMessage{
		ParticipantId: req.ParticipantId,
		Message:       fmt.Sprintf("Participant %s has left Chitty-Chat at Lamport time %d", req.ParticipantId, s.lamport_time),
		LamportTime:   s.lamport_time,
	}

	s.broadcast(leaveMessage, req.ParticipantId) //Broadcast the leaveMessage to all participants using the broadcast() method

	return &pb.LeaveResponse{ //Returns the leaveMessage to the leaving client, to know that the leave was succesful
		Message:     leaveMessage.Message,
		LamportTime: s.lamport_time,
	}, nil
}

func (s *ChittyChatService) Broadcast(msg *pb.BroadcastMessage, grpc grpc.ServerStreamingServer[pb.BroadcastMessage]) error {
	s.mu.Lock()
	msgChannel, exists := s.participants[msg.ParticipantId]
	defer s.mu.Unlock()

	if !exists {
		return fmt.Errorf("participant %s does not exists", msg.ParticipantId)
	}

	for message := range msgChannel {
		if err := grpc.Send(message); err != nil {
			return err
		}
	}

	return nil
}

func (s *ChittyChatService) broadcast(msg *pb.BroadcastMessage, joinedParticipantId string) {
	for ParticipantID, ch := range s.participants {
		if ParticipantID != joinedParticipantId {
			ch <- msg
		}
	}
}
