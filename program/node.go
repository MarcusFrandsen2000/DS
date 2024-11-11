package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "DS/proto"
)


var (
	// Hardcoded list of peer addresses
	nodeAddresses = []string{"localhost:5001", "localhost:5002", "localhost:5003"}
	// Mutex to protect critical section
	mu sync.Mutex
)

// Node represents a peer in the distributed system
type Node struct {
	ID        int32
	HasToken  bool
	NextNode  int32
	PrevNode  int32
	NodeConn  map[int32]pb.MutualExclusionServiceClient
}

func NewNode(id int32) *Node {
	n := &Node{
		ID:       id,
		HasToken: id == 1, // Node 1 starts with the token
		NodeConn: make(map[int32]pb.MutualExclusionServiceClient),
		NextNode: int32((id % int32(len(nodeAddresses))) + 1), // Determine the next node in the ring
		PrevNode: int32((id-2+int32(len(nodeAddresses))) % int32(len(nodeAddresses)) + 1), // Determine the previous node in the ring
	}
	for i, addr := range nodeAddresses {
		if int32(i+1) != n.ID {
			conn, err := grpc.Dial(addr, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("Failed to connect to node %d at %s: %v", i+1, addr, err)
			}
			n.NodeConn[int32(i+1)] = pb.NewMutualExclusionServiceClient(conn)
		}
	}
	return n
}

func (n *Node) RequestToken(ctx context.Context, req *pb.RequestAccess) (*pb.Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Node %d received request from Node %d for the token\n", n.ID, req.NodeId)

	if n.HasToken {
		fmt.Printf("Node %d has the token and will send it to Node %d\n", n.ID, req.NodeId)
		n.HasToken = false
		_, err := n.NodeConn[req.NodeId].ReceiveToken(ctx, &Token{})
		if err != nil {
			log.Printf("Failed to send token to node %d: %v", req.NodeId, err)
		}
	} else {
		fmt.Printf("Node %d does not have the token, forwarding the request\n", n.ID)
		// Forward request to next node
		_, err := n.NodeConn[n.NextNode].RequestToken(ctx, req)
		if err != nil {
			log.Printf("Failed to forward request to node %d: %v", n.NextNode, err)
		}
	}

	return &Empty{}, nil
}

func (n *Node) ReceiveToken(ctx context.Context, token *Token) (*pb.Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Node %d received the token\n", n.ID)
	n.HasToken = true

	// Critical section can be entered here
	fmt.Printf("Node %d is entering the critical section\n", n.ID)
	time.Sleep(2 * time.Second) // Simulate work in critical section
	fmt.Printf("Node %d is leaving the critical section\n", n.ID)

	// Pass the token to the next node
	fmt.Printf("Node %d is passing the token to Node %d\n", n.ID, n.NextNode)
	_, err := n.NodeConn[n.NextNode].ReceiveToken(ctx, &Token{})
	if err != nil {
		log.Printf("Failed to pass token to node %d: %v", n.NextNode, err)
	}

	return &Empty{}, nil
}

func (n *Node) JoinRing(ctx context.Context, req *JoinRequest) (*Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Node %d received request to join from Node %d", n.ID, req.NewNodeId)
	// Update the next node to point to the new node
	n.NextNode = req.NewNodeId
	fmt.Printf("Node %d now points to Node %d as NextNode", n.ID, n.NextNode)

	// Inform the new node about its predecessor
	_, err := n.NodeConn[req.NewNodeId].UpdatePrevNode(ctx, &UpdateNodeRequest{NodeId: n.ID})
	if err != nil {
		log.Printf("Failed to update predecessor for node %d: %v", req.NewNodeId, err)
	}

	return &Empty{}, nil
}

func (n *Node) LeaveRing(ctx context.Context, req *LeaveRequest) (*Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Node %d received request to leave from Node %d", n.ID, req.NodeId)
	if n.NextNode == req.NodeId {
		// Update the next node to point to the leaving node's next node
		n.NextNode = req.NewNextNodeId
		fmt.Printf("Node %d now points to Node %d as NextNode after Node %d left", n.ID, n.NextNode, req.NodeId)
	}

	return &Empty{}, nil
}

func (n *Node) UpdatePrevNode(ctx context.Context, req *UpdateNodeRequest) (*Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Node %d updating PrevNode to Node %d", n.ID, req.NodeId)
	n.PrevNode = req.NodeId

	return &Empty{}, nil
}

func main() {
	var wg sync.WaitGroup

	// Start 3 nodes concurrently
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		nodeID := int32(i)
		go func(nodeID int32) {
			defer wg.Done()
			node := NewNode(nodeID)

			// Run gRPC server
			lis, err := net.Listen("tcp", nodeAddresses[nodeID-1])
			if err != nil {
				log.Fatalf("Failed to listen on %s: %v", nodeAddresses[nodeID-1], err)
			}

			grpcServer := grpc.NewServer()
			pb.RegisterMutualExclusionServiceServer(grpcServer, node)
			fmt.Printf("Node %d is running at %s", nodeID, nodeAddresses[nodeID-1])

			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("Failed to serve: %v", err)
			}
		}(nodeID)
	}

	wg.Wait()
}
