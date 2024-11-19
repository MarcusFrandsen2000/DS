package main

import (
	"context"
    "log"
    "net"

    "google.golang.org/grpc"
	pb "DS/proto"
)

type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	highestBid int32
}

func (s *AuctionServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error){
	if req.Amount > s.highestBid {
        s.highestBid = req.Amount
    }

	return &pb.BidResponse{
        Status:     "Bid has been received",
        HighestBid: s.highestBid,
    }, nil
}

func (s *AuctionServer) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error){
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
	pb.RegisterAuctionServiceServer(grpcServer, &AuctionServer{})

	log.Printf("Auction server is running on port 50051")

	if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}