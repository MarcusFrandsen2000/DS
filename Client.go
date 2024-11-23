package main

import (
	"context"
	"log"
	"math/rand"
	"time"
	"fmt"

	pb "DS/proto"

	"google.golang.org/grpc"
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

	randomInt := rand.Intn(51) + 50

    return &Client{
        id:     id,
		money: int32(randomInt),
        client: pb.NewAuctionServiceClient(conn),
    }, nil
}

func (c *Client) placeBid(amount int32) (*pb.BidResponse, error){
	// Step 1: Create a context with a timeout of 1 second.
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    // Step 2: Ensure the context is cancelled when the function completes.
    defer cancel()

	response, err := c.client.Bid(ctx, &pb.BidRequest{
		Amount: amount,
		ClientID: c.id,
	})

	if err != nil {
		return nil, err
	}

	return response, err
}

func (c *Client) getState() (*pb.ResultResponse){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := c.client.Result(ctx, &pb.ResultRequest{})
	if err != nil {
		log.Printf("Failed to get the state of the Auction (highest bid)", err)
		return nil
	}

	return status
}

func main(){
	highestBidderCounter := 0
	auctionTimeframeLimit := int32(100)
	// Prompt user to enter a client ID
	var clientID int32
	fmt.Print("Enter the client ID (as an integer): ")

	// Use fmt.Scan() to get the user input
	_, err := fmt.Scan(&clientID)
	if err != nil {
		log.Fatalf("Error reading client ID: %v", err)
	}
	client, err := NewClient(clientID, "localhost:50051")

	if err != nil {
		log.Fatalf("Failed to create client %d", clientID)
	}

	for {
		time.Sleep(time.Duration(5) * time.Second) //Each node waits some time before requesting
		log.Printf("Client %d is requesting Auction State\n", clientID)
		status := client.getState()

		if status == nil {
			time.Sleep(time.Duration(10) * time.Second)
		} else {
			// Check if the timeframe has exceeded the limit
			if status.Timeframe > auctionTimeframeLimit {
				log.Printf("Client %d: Auction timeframe has exceeded the limit. Stopping bidding.", clientID)
				break
			}

			if status.HighestBidder == clientID {
				log.Printf("Client %d is already the highest bidder", clientID)
				highestBidderCounter++
				if highestBidderCounter == 3 {
					log.Printf("Going once")
					time.Sleep(time.Duration(2) * time.Second)
					log.Printf("Going twice")
					time.Sleep(time.Duration(2) * time.Second)
					log.Printf("SOOOOLD to Client %d for %d dollars", status.HighestBidder, status.HighestBid)
					break
				}
			} else if client.money <= status.HighestBid {
				log.Printf("Client %d doesnt have enough money to place a higher bid", clientID)
				break
			} else {
				bidResp, err := client.placeBid(status.HighestBid + 1)

				if err != nil {
					log.Fatalf("Error placing bid: %v", err)
					return
				}

				log.Printf("%sClient %d bid %d", bidResp.Message, clientID, bidResp.Bid)
				highestBidderCounter = 0
			}
		}
	}
}