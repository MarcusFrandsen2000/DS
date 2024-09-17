package main
import (
	"fmt"
	"math/rand"
	"sync"
)

var waitingGroup sync.WaitGroup

type Packet struct{
	seq int
	ack int
}

func client (serverToClient, clientToServer chan Packet){
	seq := rand.Intn(100) + 1 
	fmt.Printf("Client: Sending SYN, seq=%d\n", seq)
	clientToServer <- Packet{seq: seq}

	SYNACK := <- serverToClient
	fmt.Printf("Client: Received SYN-ACK, ack=%d, seq=%d\n", SYNACK.ack, SYNACK.seq)

	fmt.Printf("Client: Sending ACK, ack=%d, seq=%d\n", SYNACK.seq + 1, SYNACK.ack)
	clientToServer <- Packet{seq: SYNACK.ack, ack:SYNACK.seq + 1}
	waitingGroup.Done()
}
func server (serverToClient, clientToServer chan Packet){
	SYN := <- clientToServer
	fmt.Printf("Server: Received SYN, seq=%d\n", SYN.seq)


	seq := rand.Intn(300) + 1
	fmt.Printf("Server: Sending SYN-ACK, ack=%d, seq=%d\n", SYN.seq + 1, seq)
	serverToClient <- Packet{seq: seq, ack: SYN.seq + 1}
	
	ACK := <- clientToServer
	fmt.Printf("Server: Received ACK, ack=%d, seq=%d\n", ACK.ack, ACK.seq)
	if(ACK.ack == seq + 1){
		fmt.Printf("3 way handshake complete")
	} else {
		fmt.Printf("3 way handshake failed")
	}
	waitingGroup.Done()
}

func main(){
	waitingGroup.Add(2)
	serverToClient := make(chan Packet)
	clientToServer := make(chan Packet)

	go client(serverToClient, clientToServer)
	go server(serverToClient, clientToServer)

	waitingGroup.Wait()
}