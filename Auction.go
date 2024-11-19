package main

import (
	"context"
    "log"
    "net"

    "google.golang.org/grpc"
	pb "DS/proto"
)

type Auction struct {
	pb.UnimplementedAuctionServiceServer
	highestBid int32
	highestBidder int32 
	timeframe int64
}

func (s *Auction) Bid(ctx context.Context, req *pb.BidRequest, clientId int32) (*pb.BidResponse, error){
	if req.Amount > s.highestBid {
        s.highestBid = req.Amount
		s.highestBidder = req.ClientID
    }

	return &pb.BidResponse{
        Status:     true,
        HighestBid: s.highestBid,
    }, nil
}

func (s *Auction) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error){
	return &pb.ResultResponse{
		Outcome: "Current highest Bid",
		HighestBid: s.highestBid,
	}, nil
}

func main(){
	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Failed to listen to server: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, &Auction{})

	log.Printf("Auction server is running on port 50051")

	if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}