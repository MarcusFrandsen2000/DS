package main

import (
	"context"
    "log"
	"net"

    "google.golang.org/grpc"
	pb "DS/proto"
)

type BackupServer struct {
	pb.UnimplementedAuctionServiceServer
	bidders []int32
	highestBid int32
	highestBidder int32
	timeframe int32
}

func NewBackupServer() *BackupServer {
    return &BackupServer{
        bidders:      []int32{},
        highestBid:   0,
        highestBidder: 0,
        timeframe:    0,
    }
}

func (s *BackupServer) SyncAuctionState(ctx context.Context, req *pb.AuctionState) (*pb.Ack, error){
	s.highestBid = req.HighestBid
	s.highestBidder = req.HighestBidder
	s.timeframe = req.Timeframe

	exists := false
	for _, id := range s.bidders {
		if id == req.HighestBidder {
			exists = true
			break
		}
	}
	if !exists {
		s.bidders = append(s.bidders, req.HighestBidder)
	}

	log.Printf("The Backup Server has succesfully been updated with highest bid: %d by Client %d", s.highestBid, s.highestBidder)
	return &pb.Ack{Success: true}, nil
}

func main(){
	listener, err := net.Listen("tcp", ":50052")

	if err != nil {
		log.Fatalf("Failed to listen to server: %v", err)
	}

	grpcServer := grpc.NewServer()
	backupServer := NewBackupServer() 
	pb.RegisterAuctionServiceServer(grpcServer, backupServer)

	log.Printf("The Backup Server is running on port :50052")
	if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}