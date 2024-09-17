package main
import (
	"fmt"
	"sync"
	"time"
	"math/rand"
)
type Packet struct{
	syn int
	ack int
}
func client (clientToServer, serverToClient chan Packet){
	number := rand.Intn(100)+1
	packet := Packet{syn:number}
	clientToServer <- syn
}
func server (serverToClient, clientToServer chan Packet){
	ack := <- clientToServer
	number := rand.Intn(100)+1
	packet := Packet{syn: ack+1, }
}

func main (

)