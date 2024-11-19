package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	pb "DS/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/internal/status"
)

type Client struct {
	id int32
	money int32
	client pb.AuctionServiceClient
}


func NewClient(id int32, address string) (*Client, error) {
    conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
    if err != nil {
        return nil, err
    }

	randomInt := rand.Int(51) + 50

    return &Client{
        id:     id,
		money: int32(randomInt),
        client: pb.NewAuctionServiceClient(conn),
    }, nil
}

func (c *Client) placeBid(){
	// Step 1: Create a context with a timeout of 1 second.
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    // Step 2: Ensure the context is cancelled when the function completes.
    defer cancel()

	status, err := c.client.Result(ctx, &pb.ResultRequest{})

	if err != nil {
		return
	}
	if status.HighestBidder == c.id {
		log.Printf("%d is already the highest bidder", c.id)
		return
	}
	if c.money <= status.HighestBid {
		log.Printf("%d doesnt have enough money to place a higher bid", c.id)
		return
	}

	response, err := c.client.Bid(ctx, &pb.BidRequest{
		Amount: status.HighestBid + 1,
		ClientID: c.id,
	})

	if err != nil {
		log.Fatalf("Failed to place bid: %v", err)
	}

	//If the call was successful, print the response message.
	log.Println("Bid response:", response.Message, response.Bid)
}


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