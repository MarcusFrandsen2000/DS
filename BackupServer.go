package main

import (
	"context"
    "log"
	"net"
	"time"

    "google.golang.org/grpc"
	pb "DS/proto"
)

type BackupServer struct {
	pb.UnimplementedAuctionServiceServer
	bidders []int32
	highestBid int32
	highestBidder int32
	timeframe int32
	lastSignal   time.Time
    isPrimaryServer bool
}

// Signal handler
func (s *BackupServer) CheckSignal(ctx context.Context, req *pb.SignalRequest) (*pb.SignalResponse, error) {
    s.lastSignal = time.Now() // Update last heartbeat time
    return &pb.SignalResponse{}, nil
}

func NewBackupServer() *BackupServer {
    return &BackupServer{
        bidders:      []int32{},
        highestBid:   0,
        highestBidder: 0,
        timeframe:    0,
    }
}

func (s *BackupServer) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidResponse, error){
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

func (s *BackupServer) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultResponse, error){
	return &pb.ResultResponse{
		HighestBid: s.highestBid,
		HighestBidder: s.highestBidder,
		Timeframe: s.timeframe,
	}, nil
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

	// Start a goroutine to monitor signals and promote to primary if necessary
    go func() {
        for {
            time.Sleep(2 * time.Second) // Check every 2 seconds
            if time.Since(backupServer.lastSignal) > 5*time.Second {
                log.Println("No signal received from primary. Promoting backup to primary.")
                backupServer.isPrimaryServer = true

                // Close old listener and start serving as primary on port 50051
                // grpcServer.Stop()
                newListener, err := net.Listen("tcp", ":50051")
                if err != nil {
                    log.Fatalf("Failed to promote to primary: %v", err)
                }

                newGrpcServer := grpc.NewServer()
                pb.RegisterAuctionServiceServer(newGrpcServer, backupServer)
                log.Printf("Backup server promoted to primary on port :50051")
                if err := newGrpcServer.Serve(newListener); err != nil {
                    log.Fatalf("Failed to serve: %v", err)
                }
            }
        }
    }()

	log.Printf("The Backup Server is running on port :50052")
	if err := grpcServer.Serve(listener); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}