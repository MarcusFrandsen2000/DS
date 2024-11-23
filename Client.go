package main

import (
	"context"
	"log"
	"math/rand"
	"time"
	"sync"

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
		log.Fatalf("Failed to get the state of the Auction (highest bid)", err)
	}

	return status
}

func main(){
	var wg sync.WaitGroup

	for i := 1; i < 4; i++ {
		wg.Add(1)
		clientID := int32(i)
		go func(clientID int32) {
			defer wg.Done()
			client, err := NewClient(clientID, "localhost:50051")

			if err != nil {
				log.Fatalf("Failed to create client %d", clientID)
			}

			for {
                time.Sleep(time.Duration(5) * time.Second) //Each node waits some time before requesting
                log.Printf("Client %d is requesting Auction State\n", clientID)
				status := client.getState()

				if status.HighestBidder == clientID {
					log.Printf("Client %d is already the highest bidder", clientID)
				} else if client.money <= status.HighestBid {
					log.Printf("Client %d doesnt have enough money to place a higher bid", clientID)
				} else {
					bidResp, err := client.placeBid(status.HighestBid + 1)

					if err != nil {
						log.Fatalf("Error placing bid: %v", err)
						return
					}

					log.Printf("%sClient %d bid %d", bidResp.Message, clientID, bidResp.Bid)
				}
            }
		}(clientID)
	}
	wg.Wait()
}