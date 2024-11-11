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
    pb.UnimplementedMutualExclusionServiceServer
    ID       int32
    HasToken bool
    NextNode int32
    PrevNode int32
    NodeConn map[int32]pb.MutualExclusionServiceClient
}

func NewNode(id int32) *Node {
    n := &Node{
        ID:       id,
        HasToken: id == 1, // Node 1 starts with the token
        NodeConn: make(map[int32]pb.MutualExclusionServiceClient),
        NextNode: int32((id % int32(len(nodeAddresses))) + 1), // Determine the next node in the ring
        PrevNode: int32((id-2+int32(len(nodeAddresses))) % int32(len(nodeAddresses)) + 1),
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

func (n *Node) Request(ctx context.Context, req *pb.RequestAccess) (*pb.Empty, error) {
    mu.Lock()
    defer mu.Unlock()

    if n.HasToken {
        if n.ID == req.NodeId {
            // If the node is requesting the token from itself and it has the token, enter critical section
            fmt.Printf("Node %d already has the token and will enter the critical section\n", n.ID)
            n.enterCriticalSection()
			mu.Unlock()
            n.releaseToken()
        } else {
            // Grant the token to the requesting node
            fmt.Printf("Node %d has the token and will send it to Node %d\n", n.ID, req.NodeId)
            n.HasToken = false
            mu.Unlock() // Unlock before making the gRPC call
            _, err := n.NodeConn[req.NodeId].Grant(ctx, &pb.GrantAccess{NodeId: req.NodeId})
            mu.Lock()   // Re-lock after the gRPC call
            if err != nil {
                log.Printf("Failed to send token to node %d: %v", req.NodeId, err)
                n.HasToken = true // Reclaim the token if sending fails
            }
        }
    } else {
        fmt.Printf("Node %d does not have the token, forwarding the request\n", n.ID)
        mu.Unlock() // Unlock before making the gRPC call
        _, err := n.NodeConn[n.NextNode].Request(ctx, req)
        mu.Lock()   // Re-lock after the gRPC call
        if err != nil {
            log.Printf("Failed to forward request to node %d: %v", n.NextNode, err)
        }
    }

    return &pb.Empty{}, nil
}

func (n *Node) Grant(ctx context.Context, req *pb.GrantAccess) (*pb.Empty, error) {
    mu.Lock()
    defer mu.Unlock()

    fmt.Printf("Node %d received the token\n", n.ID)
    n.HasToken = true

    // Enter the critical section
    n.enterCriticalSection()
	mu.Unlock()
    n.releaseToken()

    return &pb.Empty{}, nil

}

func (n *Node) enterCriticalSection() {
    fmt.Printf("Node %d is entering the critical section\n", n.ID)
    time.Sleep(2 * time.Second) // Simulate work in critical section
    fmt.Printf("Node %d is leaving the critical section\n", n.ID)

}

func (n *Node) releaseToken() {
    mu.Lock()

    // Pass the token to the next node using the Release method
    fmt.Printf("Node %d is releasing the token to Node %d\n", n.ID, n.NextNode)
    n.HasToken = false

    mu.Unlock() // Unlock before making the gRPC call
    _, err := n.NodeConn[n.NextNode].Release(context.Background(), &pb.ReleaseAccess{NodeId: n.NextNode})
    mu.Lock()   // Re-lock after the gRPC call

    if err != nil {
        log.Printf("Failed to release token to node %d: %v. Retrying...", n.NextNode, err)
        n.HasToken = true // Reclaim the token if releasing fails
    } else {
        fmt.Printf("Node %d successfully released the token to Node %d\n", n.ID, n.NextNode)
    }
}

func (n *Node) Release(ctx context.Context, req *pb.ReleaseAccess) (*pb.Empty, error) {
    mu.Lock()
    defer mu.Unlock()

    fmt.Printf("Node %d received the token release from Node %d\n", n.ID, n.PrevNode)
    n.HasToken = true

    return &pb.Empty{}, nil
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
            fmt.Printf("Node %d is running at %s\n", nodeID, nodeAddresses[nodeID-1])

            go func() {
                if err := grpcServer.Serve(lis); err != nil {
                    log.Fatalf("Failed to serve: %v", err)
                }
            }()

            // Periodically request the token
            for {
                time.Sleep(time.Duration(5*nodeID) * time.Second) // Each node waits different time before requesting
                fmt.Printf("Node %d is requesting the token\n", nodeID)
                _, err := node.Request(context.Background(), &pb.RequestAccess{NodeId: nodeID})
                if err != nil {
                    log.Printf("Node %d failed to request token: %v", nodeID, err)
                }
            }
        }(nodeID)
    }

    wg.Wait()
}