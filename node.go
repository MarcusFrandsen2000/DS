package main

import (
	"context"
	"log"
	"time"

	pb "DS/proto"

	"google.golang.org/grpc"
)

func main(){
	conn, err := grpc.Dial("localhost: 50051", grpc.WithInsecure())

	if err != nil {
        log.Fatalf("Could not connect: %v", err)
    }

	defer conn.Close()
	node := pb.NewAuctionServiceClient(conn)

	bidResp, err := node.Bid(context.Background(), &pb.BidRequest{Amount: 100})

	if err != nil {
        log.Fatalf("Error placing bid: %v", err)
    }

	log.Printf("Bid response: %v, Highest Bid %d", bidResp.Status, bidResp.HighestBid)

	// Query for the current highest bid
	time.Sleep(time.Second * 1)
	resultResp, err := node.Result(context.Background(), &pb.ResultRequest{})

	if err != nil {
        log.Fatalf("Error getting result: %v", err)
    }

	log.Printf("Result response: %v, Highest Bid %d", resultResp.Outcome, resultResp.HighestBid)
}