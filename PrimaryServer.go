package main

import (
	"context"
    "log"
    "net"
	"time"

    "google.golang.org/grpc"
	pb "DS/proto"
)

type PrimaryServer struct {
	pb.UnimplementedAuctionServiceServer
	bidders []int32
	highestBid int32
	highestBidder int32
	timeframe int32
	backupServer pb.AuctionServiceClient
}

func NewPrimaryServer(backupServer pb.AuctionServiceClient) *PrimaryServer {
    return &PrimaryServer{
        bidders:       []int32{},
        highestBid:    50,
        highestBidder: 0,
        timeframe:     0,
        backupServer:  backupServer, // The gRPC client for communicating with BackupServer
    }
}

func (s *PrimaryServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error){
	if req.Amount > s.highestBid {
		exists := false
        for _, id := range s.bidders {
            if id == req.ClientID {
                exists = true
                break
            }
        }

        if !exists {
            s.bidders = append(s.bidders, req.ClientID)
        }

        s.highestBid = req.Amount
		s.highestBidder = req.ClientID
		s.timeframe++

		auctionState := &pb.AuctionState{
			HighestBid: s.highestBid,
			HighestBidder: s.highestBidder,
			Timeframe: s.timeframe,
		}

		_, err := s.backupServer.SyncAuctionState(context.Background(), auctionState)
		if err != nil {
			log.Println("Error syncing with backup:", err)
			return &pb.BidResponse{
				Status: false, 
				Message: "Failed to sync with backup. Bid not accepted", 
				Bid: req.Amount,
			}, err
		}

		return &pb.BidResponse{
			Status: true, 
			Message: "Bid accepted: ", 
			Bid: req.Amount,
		}, nil
    }

	return &pb.BidResponse{
		Status: false, 
		Message: "Bid not accepted. Bid too low: ", 
		Bid: req.Amount,
	}, nil
}

func (s *PrimaryServer) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error){
	return &pb.ResultResponse{
		HighestBid: s.highestBid,
		HighestBidder: s.highestBidder,
		Timeframe: s.timeframe,
	}, nil
}

func main(){
	// Start the backup server connection
    backupConn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect to backup server: %v", err)
    }
    defer backupConn.Close()

	backupClient := pb.NewAuctionServiceClient(backupConn)

	listener, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Failed to listen to server: %v", err)
	}

	grpcServer := grpc.NewServer()
	primaryServer := NewPrimaryServer(backupClient)
	pb.RegisterAuctionServiceServer(grpcServer, primaryServer)

	go func() {
        for {
            time.Sleep(2 * time.Second) // Send a heartbeat every 2 seconds
            _, err := backupClient.CheckSignal(context.Background(), &pb.SignalRequest{})
            if err != nil {
                log.Println("Backup server is unreachable:", err)
            }
        }
    }()

	log.Printf("Auction server is running on port :50051")
	if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}